#pragma once
#include <santa/abi.h>
#include <yasl/cont/seq/vector.h>

#include <cstdint>

namespace santa {
using Pair = SantaPair;
using Budget = SantaBudget;
enum class Status : int32_t {
  kSuccess = SANTA_SUCCESS,
  kInvalidInput = SANTA_INVALID_INPUT,
  kUnsatisfiable = SANTA_UNSATISFIABLE,
  kSearchLimitExceeded = SANTA_SEARCH_LIMIT_EXCEEDED,
  kResourceLimitExceeded = SANTA_RESOURCE_LIMIT_EXCEEDED,
  kInternalError = SANTA_INTERNAL_ERROR
};
class IRandom {
 public:
  virtual ~IRandom() = default;
  virtual uint64_t next() = 0;
};
class OsRandom final : public IRandom {
 public:
  uint64_t next() override;
};
struct Result {
  Status status;
  yasl::Vector<uint32_t> receivers;
  // Hall witness: these givers have fewer available receivers than their count.
  yasl::Vector<uint32_t> affected_givers;
  uint64_t attempts = 0;
};
// All indices are snapshot-local. Duplicate directed pairs are harmless.
Result draw(uint32_t n, yasl::Span<const Pair> exclusions, Budget budget,
            IRandom& random);
bool valid(uint32_t n, yasl::Span<const Pair> exclusions,
           yasl::Span<const uint32_t> receivers);
yasl::Vector<Pair> symmetric(yasl::Span<const Pair> pairs);
}  // namespace santa
