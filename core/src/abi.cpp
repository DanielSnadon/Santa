#include <santa/engine.h>

#include <algorithm>
#include <limits>
#include <new>

static_assert(sizeof(SantaPair) == 8);
static_assert(sizeof(SantaBudget) == 24);
uint32_t SANTA_CALL santa_abi_version(void) { return 1; }
int32_t SANTA_CALL santa_draw(uint32_t n, const SantaPair* pairs,
                              uint64_t pair_count, SantaBudget budget,
                              uint32_t* output, uint64_t output_count) {
  if (!output || output_count < n || (pair_count && !pairs) ||
      pair_count > std::numeric_limits<size_t>::max() / sizeof(SantaPair))
    return SANTA_INVALID_INPUT;
  try {
    santa::OsRandom random;
    const auto result = santa::draw(n, {pairs, static_cast<size_t>(pair_count)},
                                    budget, random);
    if (result.status == santa::Status::kSuccess)
      std::copy(result.receivers.begin(), result.receivers.end(), output);
    return static_cast<int32_t>(result.status);
  } catch (const std::bad_alloc&) {
    return SANTA_RESOURCE_LIMIT_EXCEEDED;
  } catch (const yasl::BadAlloc&) {
    return SANTA_RESOURCE_LIMIT_EXCEEDED;
  } catch (...) {
    return SANTA_INTERNAL_ERROR;
  }
}
