#include <santa/engine.h>
#include <chrono>
#include <iostream>
#include <random>
struct Random : santa::IRandom {
  std::mt19937_64 generator{123};
  uint64_t next() override { return generator(); }
};
int main() {
  Random random;
  for (uint32_t n : {10u,20u,50u,100u,1000u}) {
    for (uint32_t density : {0u,25u,75u}) {
      yasl::Vector<santa::Pair> pairs;
      // Keep a directed Hamiltonian cycle allowed, proving feasibility.
      for (uint32_t i=0; i<n; ++i) for (uint32_t j=0; j<n; ++j)
        if (i!=j && j!=(i+1)%n && random.next()%100 < density) pairs.push_back({i,j});
      const auto start = std::chrono::steady_clock::now();
      const auto r = santa::draw(n,pairs.span(),{100000,100000000,128*1024*1024},random);
      std::cout << n << ',' << density << ',' << static_cast<int>(r.status) << ',' << r.attempts << ','
        << std::chrono::duration_cast<std::chrono::microseconds>(std::chrono::steady_clock::now()-start).count()
        << ',' << (uint64_t{n}*n+64*n+(n<=20 && pairs.size()>uint64_t{n}*(n-1)/4 ? (uint64_t{1}<<n)*8 : 0)) << '\n';
    }
  }
}
