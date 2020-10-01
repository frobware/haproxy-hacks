#include <assert.h>
#include <errno.h>
#include <math.h>
#include <signal.h>
#include <stdarg.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/stat.h>

#include <curl/curl.h>

#define NELEMENTS(A) ((sizeof(A) / sizeof(A[0])))

enum info_tag { LONGINFO = 1, DOUBLEINFO };

union typeinfo {
  long longinfo;
  double doubleinfo;
};

struct output_field {
  const char *name;
  CURLINFO info;
  enum info_tag val_type;
  union typeinfo val;
};

static struct output_field output_fields[] = {
    {"namelookup", CURLINFO_NAMELOOKUP_TIME, DOUBLEINFO},
    {"connect", CURLINFO_CONNECT_TIME, DOUBLEINFO},
    {"app_connect", CURLINFO_APPCONNECT_TIME, DOUBLEINFO},
    {"pretransfer", CURLINFO_PRETRANSFER_TIME, DOUBLEINFO},
    {"starttransfer", CURLINFO_STARTTRANSFER_TIME, DOUBLEINFO},
    {"http_code", CURLINFO_RESPONSE_CODE, LONGINFO},
    {"port", CURLINFO_LOCAL_PORT, LONGINFO},
    {"total", CURLINFO_TOTAL_TIME, DOUBLEINFO},
    {"num_connects", CURLINFO_NUM_CONNECTS, LONGINFO},
};

volatile sig_atomic_t done = 0;
static void sigterm_handler(int signum) { done = 1; }

static int write_result;

static size_t write_cb(void *data, size_t size, size_t nmemb, void *userp) {
  size_t realsize = size * nmemb;

  if (write_result) {
    fprintf(stdout, "%*s", (int)size, (char *)data);
  }

  return realsize;
}

static void getinfo_or_die(CURL *curl, CURLINFO info, ...) {
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

static char *mprintf(const char *fmt, ...) {
  size_t size = 0;
  char *p = NULL;
  va_list ap;

  /* Determine required size. */
  va_start(ap, fmt);
  size = vsnprintf(p, size, fmt, ap);
  va_end(ap);

  size++; /* For '\0' */

  if ((p = malloc(size)) == NULL) {
    return NULL;
  }

  va_start(ap, fmt);
  size = vsnprintf(p, size, fmt, ap);
  va_end(ap);

  return p;
}

static void current_time(char *s, size_t sz, const char *strftime_fmt,
                         int *milliseconds) {
  struct tm *tm_info;
  struct timeval tv;

  gettimeofday(&tv, NULL);
  *milliseconds = lrint(tv.tv_usec / 1000.0);

  if (*milliseconds >= 1000) {
    *milliseconds -= 1000;
    tv.tv_sec++;
  }

  tm_info = localtime(&tv.tv_sec);
  strftime(s, sz, strftime_fmt, tm_info);
}

int main(int argc, char *argv[]) {
  CURL *curl_handle = NULL;
  CURLcode curl_rc;
  size_t i, j, n = -1;
  int reuse = 0;
  char time_buffer[1024];
  char *url = NULL;
  char *user_agent = NULL;
  int milliseconds;
  struct sigaction action;

  if (argc < 2) {
    fprintf(stderr, "usage: <URL>\n");
    exit(EXIT_FAILURE);
  }

  curl_global_init(CURL_GLOBAL_DEFAULT);

  if (getenv("N") != NULL) {
    n = atoi(getenv("N"));
  }

  if (getenv("W") != NULL) {
    write_result = atoi(getenv("W"));
  }

  if (getenv("R") != NULL) {
    reuse = atoi(getenv("R"));
  }

  if (getenv("O") != NULL) {
    current_time(time_buffer, sizeof(time_buffer) - 1, "%Y-%m-%d-%H%M%S",
                 &milliseconds);

    char *stdout_filename = mprintf("curler-R%d-%s.stdout", reuse, time_buffer);
    char *stderr_filename = mprintf("curler-R%d-%s.stderr", reuse, time_buffer);

    FILE *new_stdout = fopen(stdout_filename, "w");
    if (new_stdout == NULL) {
      fprintf(stderr, "failed to open \"%s\": %s\n", stdout_filename,
              strerror(errno));
      exit(EXIT_FAILURE);
    }

    FILE *new_stderr = fopen(stderr_filename, "w");
    if (new_stderr == NULL) {
      fprintf(stderr, "failed to open \"%s\": %s\n", stderr_filename,
              strerror(errno));
      exit(EXIT_FAILURE);
    }

    fprintf(stdout, "reopening stdout to \"%s\"\n", stdout_filename);
    fprintf(stdout, "reopening stderr to \"%s\"\n", stderr_filename);

    if ((stdout = freopen(stdout_filename, "w", stdout)) == NULL) {
      perror("freopen stdout");
      exit(EXIT_FAILURE);
    }

    if ((stderr = freopen(stderr_filename, "w", stderr)) == NULL) {
      perror("freopen stderr");
      exit(EXIT_FAILURE);
    }
  }

  memset(&action, 0, sizeof(action));
  action.sa_handler = sigterm_handler;
  sigaction(SIGTERM, &action, NULL);

  for (i = 0; !done && (n == -1 || i < n); i++) {
    if (url != NULL) {
      free(url);
    }

    if (user_agent != NULL) {
      free(user_agent);
    }

    if (curl_handle == NULL) {
      curl_handle = curl_easy_init();
      assert(curl_handle);
    }

    url = mprintf("%s?queryid=%zd", argv[1], i);
    user_agent = mprintf("curler/queryid=%zd", i);

    curl_easy_setopt(curl_handle, CURLOPT_URL, url);
    curl_easy_setopt(curl_handle, CURLOPT_USERAGENT, user_agent);
    curl_easy_setopt(curl_handle, CURLOPT_BUFFERSIZE, 102400L);
    curl_easy_setopt(curl_handle, CURLOPT_WRITEFUNCTION, write_cb);
    curl_easy_setopt(curl_handle, CURLOPT_SSL_VERIFYPEER, 0L);
    curl_easy_setopt(curl_handle, CURLOPT_SSL_VERIFYHOST, 0L);
    curl_easy_setopt(curl_handle, CURLOPT_FOLLOWLOCATION, 1L);
    curl_easy_setopt(curl_handle, CURLOPT_HTTP_VERSION, CURL_HTTP_VERSION_1_1);
    curl_easy_setopt(curl_handle, CURLOPT_NOPROGRESS, 1L);

    current_time(time_buffer, sizeof(time_buffer) - 1, "%H:%M:%S",
                 &milliseconds);

    /* client.Get() */
    curl_rc = curl_easy_perform(curl_handle);

    if (curl_rc != CURLE_OK) {
      fprintf(stderr, "%zd: curl_easy_perform() failed: %s (error=%zd)\n", i,
              curl_easy_strerror(curl_rc), (size_t)curl_rc);
      goto easy_perform_cleanup;
    }

    fprintf(stdout, "%zd ", i);
    fprintf(stdout, "%s.%03d ", time_buffer, milliseconds);

    for (j = 0; j < NELEMENTS(output_fields); j++) {
      switch (output_fields[j].val_type) {
      case DOUBLEINFO:
        getinfo_or_die(curl_handle, output_fields[j].info,
                       &output_fields[j].val.doubleinfo);
        fprintf(stdout, "%s %.06f", output_fields[j].name,
                output_fields[j].val.doubleinfo);
        break;
      case LONGINFO:
        getinfo_or_die(curl_handle, output_fields[j].info,
                       &output_fields[j].val.longinfo);
        fprintf(stdout, "%s %ld", output_fields[j].name,
                output_fields[j].val.longinfo);
        break;
      }

      if (j + 1 < NELEMENTS(output_fields)) {
        fprintf(stdout, " ");
      } else {
        fprintf(stdout, "\n");
      }
    }

  easy_perform_cleanup:
    if (!reuse) {
      curl_easy_cleanup(curl_handle); // End a libcurl easy handle
      curl_handle = NULL;
    }
  }

  fflush(stdout);
  fflush(stderr);

  curl_global_cleanup();

  return 0;
}
