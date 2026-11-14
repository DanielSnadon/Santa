#ifndef SANTA_ABI_H
#define SANTA_ABI_H
#include <yasl/core/types.h>
#if defined(_WIN32)
#define SANTA_API __declspec(dllexport)
#define SANTA_CALL __cdecl
#else
#define SANTA_API __attribute__((visibility("default")))
#define SANTA_CALL
#endif
#ifdef __cplusplus
extern "C" {
#endif
/* v1: cdecl, fixed-width integers, natural alignment, caller-owned buffers.
   Pairs are DIRECTED; adapters expand symmetric exclusions to both directions.
   No pointers retained. Calls are independent and thread-safe. Output untouched
   unless Success. All lengths are element counts, not bytes. Inputs/output must
   point to valid, non-overlapping memory. Cancellation is not exposed in v1. */
typedef struct SantaPair {
  uint32 giver;
  uint32 receiver;
} SantaPair;
typedef struct SantaBudget {
  uint64 max_attempts;
  uint64 max_work;
  uint64 max_memory_bytes;
} SantaBudget;
enum SantaStatus {
  SANTA_SUCCESS = 0,
  SANTA_INVALID_INPUT = 1,
  SANTA_UNSATISFIABLE = 2,
  SANTA_SEARCH_LIMIT_EXCEEDED = 3,
  SANTA_CANCELLED = 4,
  SANTA_INTERNAL_ERROR = 5,
  SANTA_RESOURCE_LIMIT_EXCEEDED = 6
};
SANTA_API uint32 SANTA_CALL santa_abi_version(void);
SANTA_API int32 SANTA_CALL santa_draw(uint32 n, const SantaPair* pairs,
                                        uint64 pair_count, SantaBudget budget,
                                        uint32* output, uint64 output_count);
#ifdef __cplusplus
}
#endif
#endif
