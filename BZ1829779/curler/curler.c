#include <curl/curl.h>
#include <math.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#undef NDEBUG
#include <assert.h>

static size_t write_cb(void *data, size_t size, size_t nmemb, void *userp) {
  size_t realsize = size * nmemb;
  return realsize;
}

int main(int argc, char *argv[]) {
  CURL *curl_handle;
  CURLcode res;
  int i, n = 10;

  if (argc < 2) {
    fprintf(stderr, "usage: <host>\n");
    exit(EXIT_FAILURE);
  }

  curl_global_init(CURL_GLOBAL_DEFAULT);

  if (getenv("N") != NULL) {
    n = atoi(getenv("N"));
  }

  char urlbuf[8192];
  assert(strlen(argv[1]) < 8000);

  for (i = 0; i < n; i++) {
    curl_handle = curl_easy_init();
    assert(curl_handle);

    if (curl_handle) {
      long longinfo;
      double doubleinfo;

      sprintf(urlbuf, "%s/?queryid=%d", argv[1], i+1);
      curl_easy_setopt(curl_handle, CURLOPT_URL, urlbuf);
      curl_easy_setopt(curl_handle, CURLOPT_WRITEFUNCTION, write_cb);

      curl_easy_setopt(curl_handle, CURLOPT_SSL_VERIFYPEER, 0L);
      curl_easy_setopt(curl_handle, CURLOPT_SSL_VERIFYHOST, 0L);

      fprintf(stdout, "%d ", i+1);

#ifndef NOTIMESTAMP
      {
        char buffer[26];
        int millisec;
        struct tm *tm_info;
        struct timeval tv;

        gettimeofday(&tv, NULL);
        millisec = lrint(tv.tv_usec / 1000.0);
        if (millisec >= 1000) {
          millisec -= 1000;
          tv.tv_sec++;
        }

        tm_info = localtime(&tv.tv_sec);
        strftime(buffer, 26, "%H:%M:%S", tm_info);
        fprintf(stdout, "%s.%03d ", buffer, millisec);
      }
#endif

      /* client.Get() */
      res = curl_easy_perform(curl_handle);

      if (res != CURLE_OK) {
        fprintf(stdout, "curl_easy_perform() failed: %s\n",
                curl_easy_strerror(res));
      } else {
        assert((CURLE_OK == curl_easy_getinfo(curl_handle,
                                              CURLINFO_NAMELOOKUP_TIME,
                                              &doubleinfo)));
        fprintf(stdout, "namelookup %0.6f ", doubleinfo);

        assert(
            (CURLE_OK == curl_easy_getinfo(curl_handle, CURLINFO_CONNECT_TIME,
                                           &doubleinfo)));
        fprintf(stdout, "connect %0.6f ", doubleinfo);

        assert((CURLE_OK == curl_easy_getinfo(curl_handle,
                                              CURLINFO_PRETRANSFER_TIME,
                                              &doubleinfo)));
        fprintf(stdout, "pre %0.6f ", doubleinfo);

        assert((CURLE_OK == curl_easy_getinfo(curl_handle,
                                              CURLINFO_STARTTRANSFER_TIME,
                                              &doubleinfo)));
        fprintf(stdout, "start %0.6f ", doubleinfo);

        assert(
            (CURLE_OK == curl_easy_getinfo(curl_handle, CURLINFO_RESPONSE_CODE,
                                           &longinfo)));
        fprintf(stdout, "http_code %03ld ", longinfo);

        assert((CURLE_OK == curl_easy_getinfo(curl_handle, CURLINFO_LOCAL_PORT,
                                              &longinfo)));
        fprintf(stdout, "port %ld ", longinfo);

        assert((CURLE_OK == curl_easy_getinfo(curl_handle, CURLINFO_TOTAL_TIME,
                                              &doubleinfo)));
        fprintf(stdout, "total %0.6f\n", doubleinfo);
      }

      curl_easy_cleanup(curl_handle); // End a libcurl easy handle
    }
  }

  curl_global_cleanup();

  return 0;
}
