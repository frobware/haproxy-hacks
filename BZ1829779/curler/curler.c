#include <curl/curl.h>
#include <math.h>
#include <stdarg.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#undef NDEBUG
#include <assert.h>

static size_t write_cb(void *data, size_t size, size_t nmemb, void *userp) {
  size_t realsize = size * nmemb;
  return realsize;
}

void getinfo_or_die(CURL *curl, CURLINFO info, ...) {
  va_list arg;
  void *paramp;
  CURLcode result;

  va_start(arg, info);
  paramp = va_arg(arg, void *);
  result = curl_easy_getinfo(curl, info, paramp);
  va_end(arg);

  if (result != CURLE_OK) {
    abort();
  }
}

int main(int argc, char *argv[]) {
  CURL *curl_handle = NULL;
  CURLcode res;
  int i, n = 10, reuse = 1;
  char buffer[26];
  int millisec;
  struct tm *tm_info;
  struct timeval tv;
  char urlbuf[8192];

  if (argc < 2) {
    fprintf(stderr, "usage: <host>\n");
    exit(EXIT_FAILURE);
  }

  curl_global_init(CURL_GLOBAL_DEFAULT);

  if (getenv("N") != NULL) {
    n = atoi(getenv("N"));
  }

  if (getenv("R") != NULL) {
    reuse = atoi(getenv("R"));
  }

  for (i = 0; i <= n; i++) {
    if (curl_handle == NULL) {
      curl_handle = curl_easy_init();
      assert(curl_handle);
    }

    long longinfo;
    double doubleinfo;

    sprintf(urlbuf, "%s/?queryid=%d", argv[1], i);
    curl_easy_setopt(curl_handle, CURLOPT_URL, urlbuf);
    curl_easy_setopt(curl_handle, CURLOPT_WRITEFUNCTION, write_cb);

    curl_easy_setopt(curl_handle, CURLOPT_SSL_VERIFYPEER, 0L);
    curl_easy_setopt(curl_handle, CURLOPT_SSL_VERIFYHOST, 0L);

    gettimeofday(&tv, NULL);
    millisec = lrint(tv.tv_usec / 1000.0);
    if (millisec >= 1000) {
      millisec -= 1000;
      tv.tv_sec++;
    }

    tm_info = localtime(&tv.tv_sec);
    strftime(buffer, 26, "%H:%M:%S", tm_info);

    curl_easy_setopt(curl_handle, CURLOPT_FOLLOWLOCATION, 1L);

    /* client.Get() */
    res = curl_easy_perform(curl_handle);

    if (res != CURLE_OK) {
      fprintf(stderr, "%ld: curl_easy_perform() failed: %s (error=%zd)\n",
              i+1, curl_easy_strerror(res), (size_t)res);
      goto out;
    }

    if (i == 0) {
      /* absorb cost of namelookup, process loading, et al. First
	 result always skews the results. */
      continue;
    }

      fprintf(stdout, "%d ", i+1);
	fprintf(stdout, "%s.%03d ", buffer, millisec);

    getinfo_or_die(curl_handle, CURLINFO_NAMELOOKUP_TIME, &doubleinfo);
    fprintf(stdout, "namelookup %0.6f ", doubleinfo);

    getinfo_or_die(curl_handle, CURLINFO_CONNECT_TIME, &doubleinfo);
    fprintf(stdout, "connect %0.6f ", doubleinfo);

    getinfo_or_die(curl_handle, CURLINFO_APPCONNECT_TIME, &doubleinfo);
    fprintf(stdout, "app_connect %0.6f ", doubleinfo);

    getinfo_or_die(curl_handle, CURLINFO_PRETRANSFER_TIME, &doubleinfo);
    fprintf(stdout, "pretransfer %0.6f ", doubleinfo);

    getinfo_or_die(curl_handle, CURLINFO_STARTTRANSFER_TIME, &doubleinfo);
    fprintf(stdout, "starttransfer %0.6f ", doubleinfo);

    getinfo_or_die(curl_handle, CURLINFO_RESPONSE_CODE, &longinfo);
    fprintf(stdout, "http_code %03ld ", longinfo);

    getinfo_or_die(curl_handle, CURLINFO_LOCAL_PORT, &longinfo);
    fprintf(stdout, "port %ld ", longinfo);

    getinfo_or_die(curl_handle, CURLINFO_TOTAL_TIME, &doubleinfo);
    fprintf(stdout, "total %0.6f\n", doubleinfo);

  out:
    if (!reuse) {
      curl_easy_cleanup(curl_handle); // End a libcurl easy handle
      curl_handle = NULL;
    }
  }

  curl_global_cleanup();

  return 0;
}
