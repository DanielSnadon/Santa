FROM mcr.microsoft.com/dotnet/sdk:10.0 AS build
RUN apt-get update && apt-get install -y --no-install-recommends cmake g++ make && rm -rf /var/lib/apt/lists/*
WORKDIR /src
COPY . .
RUN cmake -S . -B build -DCMAKE_BUILD_TYPE=Release -DBUILD_TESTING=OFF && cmake --build build --target santa_native -j 2
RUN dotnet restore santa/santa.csproj --locked-mode && dotnet publish santa/santa.csproj -c Release --no-restore -o /out
FROM mcr.microsoft.com/dotnet/aspnet:10.0
WORKDIR /app
USER root
RUN apt-get -o Acquire::Retries=2 -o Acquire::http::Timeout=30 update && apt-get -o Acquire::Retries=2 -o Acquire::http::Timeout=30 install -y --no-install-recommends libgssapi-krb5-2 && rm -rf /var/lib/apt/lists/* \
    && mkdir -p /var/lib/santa/keys && chown -R app:app /var/lib/santa
COPY --from=build /out .
USER app
ENV ASPNETCORE_HTTP_PORTS=8080
ENV DataProtection__Path=/var/lib/santa/keys
EXPOSE 8080
ENTRYPOINT ["dotnet", "santa.dll"]
