#include <assert.h>
#include <santa/abi.h>
#include <stdint.h>

int main(void) {
  SantaBudget budget = {10000, 1000000, 1048576};
  uint32_t output[2] = {99, 99};
  if (sizeof(SantaPair) != 8 || sizeof(SantaBudget) != 24 ||
      santa_abi_version() != 1)
    return 1;
  assert(sizeof(uint32_t) == 32);
  if (santa_draw(2, 0, 0, budget, output, 2) != SANTA_SUCCESS)
    return 2;
  return output[0] != 1 || output[1] != 0;
}
