/* malloc/calloc interceptor; adjust size to balloon allocations. */

#define _GNU_SOURCE
#include <dlfcn.h>
#include <memory.h>
#include <stdio.h>
#include <assert.h>

#define ONE_MB 1048576 /* 1<<20 */

#ifndef MALLOC_BALLOON
#define MALLOC_BALLOON ONE_MB
#endif

#ifndef CALLOC_BALLOON
#define CALLOC_BALLOON 0
#endif

static void *(*real_malloc)(size_t size);
static void *(*real_calloc)(size_t n, size_t size);
static void *(*real_free)(void *ptr);

static int malloc_balloon = MALLOC_BALLOON;
static int calloc_balloon = CALLOC_BALLOON;

static size_t calloc_buffer_alloc = 0;
static unsigned char calloc_buffer[1<<20];

static void malloc_init(void) {
  if (real_malloc == NULL) {
    real_malloc = dlsym(RTLD_NEXT, "malloc");
  }
  assert(real_malloc);
}

static void calloc_init(void) {
  if (real_calloc == NULL) {
    real_calloc = dlsym(RTLD_NEXT, "calloc");
  }
  assert(real_calloc);
}

static void free_init(void) {
  if (real_free == NULL) {
    real_free = dlsym(RTLD_NEXT, "free");
  }
  assert(real_free);
}

void *malloc(size_t size) {
  void *ptr;

  if (real_malloc == NULL) {
    malloc_init();
  }

  if (real_calloc == NULL) {
    calloc_init();
  }
  
  if (size > 0) {
    size += malloc_balloon;
  }

  ptr = real_malloc(size);

#if 0
  fprintf(stderr, "malloc %zd == %p\n", size, ptr);
#endif

  return ptr;
}

void *calloc(size_t n, size_t size) {
  void *ptr;

  if (real_calloc == NULL) {
    /* one time only because dlsym calls calloc. Force the first call
       to calloc to come from a static buffer */
    calloc_buffer_alloc++;
    assert(calloc_buffer_alloc == 1);
    return calloc_buffer;
  }

  if (size > 0) {
    size += calloc_balloon;
  }

  ptr = real_calloc(n, size);
#if 0
  fprintf(stderr, "calloc n=%10zd size=%8zd == %p\n", n, size, ptr);
#endif
  return ptr;
}

void free(void *p) {
  if ((unsigned char *)p == calloc_buffer) {
    fprintf(stderr, "ignoring free in calloc buffer\n");
    return;
  }
  free_init();
  real_free(p);
}
