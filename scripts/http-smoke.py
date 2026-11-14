#!/usr/bin/env python3
"""Real HTTP regression checks. Uses unique accounts; no database cleanup."""
import http.cookiejar
import json
import re
import urllib.request
import urllib.parse
import urllib.error
import uuid
import os

BASE = os.environ.get('SANTA_TEST_URL','http://127.0.0.1:5080')
class Client:
    def __init__(self):
        self.opener = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(http.cookiejar.CookieJar()))
    def request(self,path,body=None,method=None,csrf=True,form=False):
        headers={}
        if body is not None:
            if form:
                body=urllib.parse.urlencode(body).encode();headers['Content-Type']='application/x-www-form-urlencoded'
            else:
                body=json.dumps(body).encode();headers['Content-Type']='application/json'
                if csrf:
                    _,session=self.request('/api/session')
                    headers['X-CSRF-TOKEN']=json.loads(session)['csrf']
        request=urllib.request.Request(BASE+path,body,headers,method=method)
        try:
            with self.opener.open(request) as response: return response.status,response.read().decode()
        except urllib.error.HTTPError as e: return e.code,e.read().decode()
    def register(self,email,name):
        status,page=self.request('/register');assert status==200
        token=re.search(r'name="__RequestVerificationToken" value="([^"]+)"',page)[1]
        status,page=self.request('/account/register',{'email':email,'name':name,'password':'Test-Gift-Password42!','__RequestVerificationToken':token},form=True)
        assert status==200 and 'Новая игра' in page,(status,page[:500])

clients=[Client() for _ in range(4)]
suffix=uuid.uuid4().hex[:10]
emails=[f'http-{suffix}-{i}@example.test' for i in range(4)]
for i,c in enumerate(clients):c.register(emails[i],f'Guest {i}')
owner,a,b,outsider=clients
anon=Client()
assert anon.request('/api/games/')[0]==401
assert owner.request('/api/games/',{'title':'missing csrf'},csrf=False)[0]==400
status,data=owner.request('/api/games/',{'title':'HTTP test'})
assert status==200,(status,data)
game=json.loads(data)['id'];path='/api/games/'+game
status,data=owner.request(path+'/configuration',{'version':1,'emails':emails[:3],'exclusions':[]},method='PUT');assert status==204,(status,data)
assert outsider.request(path)[0]==404
assert a.request(path+'/configuration')[0]==403
assert a.request(path+'/publish',{'version':2,'requestKey':str(uuid.uuid4())})[0]==403
key=str(uuid.uuid4())
status,data=owner.request(path+'/publish',{'version':2,'requestKey':key});assert status==200,(status,data)
assert owner.request(path+'/publish',{'version':2,'requestKey':key})[0]==200
assert owner.request(path+'/publish',{'version':2,'requestKey':str(uuid.uuid4())})[0]==409
assert owner.request(path+'/configuration',{'version':2,'emails':emails,'exclusions':[]},method='PUT')[0]==409
recipients=[]
for i,c in enumerate(clients[:3]):
    status,data=c.request(path+'/my-assignment');assert status==200,(status,data)
    value=json.loads(data);assert set(value)=={'name','email'} and value['email']!=emails[i]
    recipients.append(value['email'])
    status,page=c.request('/games/'+game)
    assert status==200 and value['email'] in page
    # Published HTML contains the one recipient; no full configuration or result.
    for other in emails[:3]:
        if other!=value['email']:assert other not in page
assert set(recipients)==set(emails[:3])
assert outsider.request(path+'/my-assignment')[0]==404
# A client-supplied giver doesn't override the authenticated identity.
status,data=a.request(path+'/my-assignment?giverId='+uuid.uuid4().hex)
assert status==200 and json.loads(data)['email']==recipients[1]
assert owner.request(path+'/close',{})[0]==204
assert a.request(path+'/my-assignment')[0]==404
# Login/logout and form CSRF.
status,page=owner.request('/')
token=re.search(r'name="__RequestVerificationToken" value="([^"]+)"',page)[1]
assert owner.request('/account/logout',{'__RequestVerificationToken':token},form=True)[0]==200
assert owner.request('/api/games/')[0]==401
status,page=owner.request('/login');token=re.search(r'name="__RequestVerificationToken" value="([^"]+)"',page)[1]
status,page=owner.request('/account/login',{'email':emails[0],'password':'Test-Gift-Password42!','__RequestVerificationToken':token},form=True)
assert status==200 and 'Новая игра' in page
print('PASS: registration, cookie login/logout, CSRF, game flow, cross-account permissions, own result only, retries, close')
