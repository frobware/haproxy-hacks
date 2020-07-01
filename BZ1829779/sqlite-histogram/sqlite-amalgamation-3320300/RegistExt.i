# 1 "RegistExt.cpp"
# 1 "<built-in>"
# 1 "<command-line>"
# 1 "/usr/include/stdc-predef.h" 1 3 4
# 1 "<command-line>" 2
# 1 "RegistExt.cpp"
# 67 "RegistExt.cpp"
# 1 "RegistExt.h" 1
# 28 "RegistExt.h"
# 1 "sqlite3ext.h" 1
# 20 "sqlite3ext.h"
# 1 "sqlite3.h" 1
# 35 "sqlite3.h"
# 1 "/usr/lib/gcc/x86_64-redhat-linux/10/include/stdarg.h" 1 3 4
# 40 "/usr/lib/gcc/x86_64-redhat-linux/10/include/stdarg.h" 3 4

# 40 "/usr/lib/gcc/x86_64-redhat-linux/10/include/stdarg.h" 3 4
typedef __builtin_va_list __gnuc_va_list;
# 99 "/usr/lib/gcc/x86_64-redhat-linux/10/include/stdarg.h" 3 4
typedef __gnuc_va_list va_list;
# 36 "sqlite3.h" 2






# 41 "sqlite3.h"
extern "C" {
# 162 "sqlite3.h"
 extern const char sqlite3_version[];
 const char *sqlite3_libversion(void);
 const char *sqlite3_sourceid(void);
 int sqlite3_libversion_number(void);
# 190 "sqlite3.h"
 int sqlite3_compileoption_used(const char *zOptName);
 const char *sqlite3_compileoption_get(int N);
# 233 "sqlite3.h"
 int sqlite3_threadsafe(void);
# 249 "sqlite3.h"
typedef struct sqlite3 sqlite3;
# 278 "sqlite3.h"
  typedef long long int sqlite_int64;
  typedef unsigned long long int sqlite_uint64;

typedef sqlite_int64 sqlite3_int64;
typedef sqlite_uint64 sqlite3_uint64;
# 330 "sqlite3.h"
 int sqlite3_close(sqlite3*);
 int sqlite3_close_v2(sqlite3*);






typedef int (*sqlite3_callback)(void*,int,char**, char**);
# 402 "sqlite3.h"
 int sqlite3_exec(
  sqlite3*,
  const char *sql,
  int (*callback)(void*,int,char**,char**),
  void *,
  char **errmsg
);
# 680 "sqlite3.h"
typedef struct sqlite3_file sqlite3_file;
struct sqlite3_file {
  const struct sqlite3_io_methods *pMethods;
};
# 779 "sqlite3.h"
typedef struct sqlite3_io_methods sqlite3_io_methods;
struct sqlite3_io_methods {
  int iVersion;
  int (*xClose)(sqlite3_file*);
  int (*xRead)(sqlite3_file*, void*, int iAmt, sqlite3_int64 iOfst);
  int (*xWrite)(sqlite3_file*, const void*, int iAmt, sqlite3_int64 iOfst);
  int (*xTruncate)(sqlite3_file*, sqlite3_int64 size);
  int (*xSync)(sqlite3_file*, int flags);
  int (*xFileSize)(sqlite3_file*, sqlite3_int64 *pSize);
  int (*xLock)(sqlite3_file*, int);
  int (*xUnlock)(sqlite3_file*, int);
  int (*xCheckReservedLock)(sqlite3_file*, int *pResOut);
  int (*xFileControl)(sqlite3_file*, int op, void *pArg);
  int (*xSectorSize)(sqlite3_file*);
  int (*xDeviceCharacteristics)(sqlite3_file*);

  int (*xShmMap)(sqlite3_file*, int iPg, int pgsz, int, void volatile**);
  int (*xShmLock)(sqlite3_file*, int offset, int n, int flags);
  void (*xShmBarrier)(sqlite3_file*);
  int (*xShmUnmap)(sqlite3_file*, int deleteFlag);

  int (*xFetch)(sqlite3_file*, sqlite3_int64 iOfst, int iAmt, void **pp);
  int (*xUnfetch)(sqlite3_file*, sqlite3_int64 iOfst, void *p);


};
# 1183 "sqlite3.h"
typedef struct sqlite3_mutex sqlite3_mutex;
# 1193 "sqlite3.h"
typedef struct sqlite3_api_routines sqlite3_api_routines;
# 1364 "sqlite3.h"
typedef struct sqlite3_vfs sqlite3_vfs;
typedef void (*sqlite3_syscall_ptr)(void);
struct sqlite3_vfs {
  int iVersion;
  int szOsFile;
  int mxPathname;
  sqlite3_vfs *pNext;
  const char *zName;
  void *pAppData;
  int (*xOpen)(sqlite3_vfs*, const char *zName, sqlite3_file*,
               int flags, int *pOutFlags);
  int (*xDelete)(sqlite3_vfs*, const char *zName, int syncDir);
  int (*xAccess)(sqlite3_vfs*, const char *zName, int flags, int *pResOut);
  int (*xFullPathname)(sqlite3_vfs*, const char *zName, int nOut, char *zOut);
  void *(*xDlOpen)(sqlite3_vfs*, const char *zFilename);
  void (*xDlError)(sqlite3_vfs*, int nByte, char *zErrMsg);
  void (*(*xDlSym)(sqlite3_vfs*,void*, const char *zSymbol))(void);
  void (*xDlClose)(sqlite3_vfs*, void*);
  int (*xRandomness)(sqlite3_vfs*, int nByte, char *zOut);
  int (*xSleep)(sqlite3_vfs*, int microseconds);
  int (*xCurrentTime)(sqlite3_vfs*, double*);
  int (*xGetLastError)(sqlite3_vfs*, int, char *);




  int (*xCurrentTimeInt64)(sqlite3_vfs*, sqlite3_int64*);




  int (*xSetSystemCall)(sqlite3_vfs*, const char *zName, sqlite3_syscall_ptr);
  sqlite3_syscall_ptr (*xGetSystemCall)(sqlite3_vfs*, const char *zName);
  const char *(*xNextSystemCall)(sqlite3_vfs*, const char *zName);





};
# 1542 "sqlite3.h"
 int sqlite3_initialize(void);
 int sqlite3_shutdown(void);
 int sqlite3_os_init(void);
 int sqlite3_os_end(void);
# 1578 "sqlite3.h"
 int sqlite3_config(int, ...);
# 1597 "sqlite3.h"
 int sqlite3_db_config(sqlite3*, int op, ...);
# 1662 "sqlite3.h"
typedef struct sqlite3_mem_methods sqlite3_mem_methods;
struct sqlite3_mem_methods {
  void *(*xMalloc)(int);
  void (*xFree)(void*);
  void *(*xRealloc)(void*,int);
  int (*xSize)(void*);
  int (*xRoundup)(int);
  int (*xInit)(void*);
  void (*xShutdown)(void*);
  void *pAppData;
};
# 2355 "sqlite3.h"
 int sqlite3_extended_result_codes(sqlite3*, int onoff);
# 2417 "sqlite3.h"
 sqlite3_int64 sqlite3_last_insert_rowid(sqlite3*);
# 2427 "sqlite3.h"
 void sqlite3_set_last_insert_rowid(sqlite3*,sqlite3_int64);
# 2485 "sqlite3.h"
 int sqlite3_changes(sqlite3*);
# 2522 "sqlite3.h"
 int sqlite3_total_changes(sqlite3*);
# 2559 "sqlite3.h"
 void sqlite3_interrupt(sqlite3*);
# 2594 "sqlite3.h"
 int sqlite3_complete(const char *sql);
 int sqlite3_complete16(const void *sql);
# 2656 "sqlite3.h"
 int sqlite3_busy_handler(sqlite3*,int(*)(void*,int),void*);
# 2679 "sqlite3.h"
 int sqlite3_busy_timeout(sqlite3*, int ms);
# 2754 "sqlite3.h"
 int sqlite3_get_table(
  sqlite3 *db,
  const char *zSql,
  char ***pazResult,
  int *pnRow,
  int *pnColumn,
  char **pzErrmsg
);
 void sqlite3_free_table(char **result);
# 2804 "sqlite3.h"
 char *sqlite3_mprintf(const char*,...);
 char *sqlite3_vmprintf(const char*, va_list);
 char *sqlite3_snprintf(int,char*,const char*, ...);
 char *sqlite3_vsnprintf(int,char*,const char*, va_list);
# 2884 "sqlite3.h"
 void *sqlite3_malloc(int);
 void *sqlite3_malloc64(sqlite3_uint64);
 void *sqlite3_realloc(void*, int);
 void *sqlite3_realloc64(void*, sqlite3_uint64);
 void sqlite3_free(void*);
 sqlite3_uint64 sqlite3_msize(void*);
# 2914 "sqlite3.h"
 sqlite3_int64 sqlite3_memory_used(void);
 sqlite3_int64 sqlite3_memory_highwater(int resetFlag);
# 2938 "sqlite3.h"
 void sqlite3_randomness(int N, void *P);
# 3029 "sqlite3.h"
 int sqlite3_set_authorizer(
  sqlite3*,
  int (*xAuth)(void*,int,const char*,const char*,const char*,const char*),
  void *pUserData
);
# 3137 "sqlite3.h"
 void *sqlite3_trace(sqlite3*,
   void(*xTrace)(void*,const char*), void*);
 void *sqlite3_profile(sqlite3*,
   void(*xProfile)(void*,const char*,sqlite3_uint64), void*);
# 3228 "sqlite3.h"
 int sqlite3_trace_v2(
  sqlite3*,
  unsigned uMask,
  int(*xCallback)(unsigned,void*,void*,void*),
  void *pCtx
);
# 3267 "sqlite3.h"
 void sqlite3_progress_handler(sqlite3*, int, int(*)(void*), void*);
# 3522 "sqlite3.h"
 int sqlite3_open(
  const char *filename,
  sqlite3 **ppDb
);
 int sqlite3_open16(
  const void *filename,
  sqlite3 **ppDb
);
 int sqlite3_open_v2(
  const char *filename,
  sqlite3 **ppDb,
  int flags,
  const char *zVfs
);
# 3603 "sqlite3.h"
 const char *sqlite3_uri_parameter(const char *zFilename, const char *zParam);
 int sqlite3_uri_boolean(const char *zFile, const char *zParam, int bDefault);
 sqlite3_int64 sqlite3_uri_int64(const char*, const char*, sqlite3_int64);
 const char *sqlite3_uri_key(const char *zFilename, int N);
# 3635 "sqlite3.h"
 const char *sqlite3_filename_database(const char*);
 const char *sqlite3_filename_journal(const char*);
 const char *sqlite3_filename_wal(const char*);
# 3656 "sqlite3.h"
 sqlite3_file *sqlite3_database_file_object(const char*);
# 3703 "sqlite3.h"
 char *sqlite3_create_filename(
  const char *zDatabase,
  const char *zJournal,
  const char *zWal,
  int nParam,
  const char **azParam
);
 void sqlite3_free_filename(char*);
# 3764 "sqlite3.h"
 int sqlite3_errcode(sqlite3 *db);
 int sqlite3_extended_errcode(sqlite3 *db);
 const char *sqlite3_errmsg(sqlite3*);
 const void *sqlite3_errmsg16(sqlite3*);
 const char *sqlite3_errstr(int);
# 3794 "sqlite3.h"
typedef struct sqlite3_stmt sqlite3_stmt;
# 3836 "sqlite3.h"
 int sqlite3_limit(sqlite3*, int id, int newVal);
# 4046 "sqlite3.h"
 int sqlite3_prepare(
  sqlite3 *db,
  const char *zSql,
  int nByte,
  sqlite3_stmt **ppStmt,
  const char **pzTail
);
 int sqlite3_prepare_v2(
  sqlite3 *db,
  const char *zSql,
  int nByte,
  sqlite3_stmt **ppStmt,
  const char **pzTail
);
 int sqlite3_prepare_v3(
  sqlite3 *db,
  const char *zSql,
  int nByte,
  unsigned int prepFlags,
  sqlite3_stmt **ppStmt,
  const char **pzTail
);
 int sqlite3_prepare16(
  sqlite3 *db,
  const void *zSql,
  int nByte,
  sqlite3_stmt **ppStmt,
  const void **pzTail
);
 int sqlite3_prepare16_v2(
  sqlite3 *db,
  const void *zSql,
  int nByte,
  sqlite3_stmt **ppStmt,
  const void **pzTail
);
 int sqlite3_prepare16_v3(
  sqlite3 *db,
  const void *zSql,
  int nByte,
  unsigned int prepFlags,
  sqlite3_stmt **ppStmt,
  const void **pzTail
);
# 4129 "sqlite3.h"
 const char *sqlite3_sql(sqlite3_stmt *pStmt);
 char *sqlite3_expanded_sql(sqlite3_stmt *pStmt);
 const char *sqlite3_normalized_sql(sqlite3_stmt *pStmt);
# 4167 "sqlite3.h"
 int sqlite3_stmt_readonly(sqlite3_stmt *pStmt);
# 4179 "sqlite3.h"
 int sqlite3_stmt_isexplain(sqlite3_stmt *pStmt);
# 4200 "sqlite3.h"
 int sqlite3_stmt_busy(sqlite3_stmt*);
# 4242 "sqlite3.h"
typedef struct sqlite3_value sqlite3_value;
# 4256 "sqlite3.h"
typedef struct sqlite3_context sqlite3_context;
# 4394 "sqlite3.h"
 int sqlite3_bind_blob(sqlite3_stmt*, int, const void*, int n, void(*)(void*));
 int sqlite3_bind_blob64(sqlite3_stmt*, int, const void*, sqlite3_uint64,
                        void(*)(void*));
 int sqlite3_bind_double(sqlite3_stmt*, int, double);
 int sqlite3_bind_int(sqlite3_stmt*, int, int);
 int sqlite3_bind_int64(sqlite3_stmt*, int, sqlite3_int64);
 int sqlite3_bind_null(sqlite3_stmt*, int);
 int sqlite3_bind_text(sqlite3_stmt*,int,const char*,int,void(*)(void*));
 int sqlite3_bind_text16(sqlite3_stmt*, int, const void*, int, void(*)(void*));
 int sqlite3_bind_text64(sqlite3_stmt*, int, const char*, sqlite3_uint64,
                         void(*)(void*), unsigned char encoding);
 int sqlite3_bind_value(sqlite3_stmt*, int, const sqlite3_value*);
 int sqlite3_bind_pointer(sqlite3_stmt*, int, void*, const char*,void(*)(void*));
 int sqlite3_bind_zeroblob(sqlite3_stmt*, int, int n);
 int sqlite3_bind_zeroblob64(sqlite3_stmt*, int, sqlite3_uint64);
# 4429 "sqlite3.h"
 int sqlite3_bind_parameter_count(sqlite3_stmt*);
# 4457 "sqlite3.h"
 const char *sqlite3_bind_parameter_name(sqlite3_stmt*, int);
# 4475 "sqlite3.h"
 int sqlite3_bind_parameter_index(sqlite3_stmt*, const char *zName);
# 4485 "sqlite3.h"
 int sqlite3_clear_bindings(sqlite3_stmt*);
# 4501 "sqlite3.h"
 int sqlite3_column_count(sqlite3_stmt *pStmt);
# 4530 "sqlite3.h"
 const char *sqlite3_column_name(sqlite3_stmt*, int N);
 const void *sqlite3_column_name16(sqlite3_stmt*, int N);
# 4575 "sqlite3.h"
 const char *sqlite3_column_database_name(sqlite3_stmt*,int);
 const void *sqlite3_column_database_name16(sqlite3_stmt*,int);
 const char *sqlite3_column_table_name(sqlite3_stmt*,int);
 const void *sqlite3_column_table_name16(sqlite3_stmt*,int);
 const char *sqlite3_column_origin_name(sqlite3_stmt*,int);
 const void *sqlite3_column_origin_name16(sqlite3_stmt*,int);
# 4612 "sqlite3.h"
 const char *sqlite3_column_decltype(sqlite3_stmt*,int);
 const void *sqlite3_column_decltype16(sqlite3_stmt*,int);
# 4697 "sqlite3.h"
 int sqlite3_step(sqlite3_stmt*);
# 4718 "sqlite3.h"
 int sqlite3_data_count(sqlite3_stmt *pStmt);
# 4961 "sqlite3.h"
 const void *sqlite3_column_blob(sqlite3_stmt*, int iCol);
 double sqlite3_column_double(sqlite3_stmt*, int iCol);
 int sqlite3_column_int(sqlite3_stmt*, int iCol);
 sqlite3_int64 sqlite3_column_int64(sqlite3_stmt*, int iCol);
 const unsigned char *sqlite3_column_text(sqlite3_stmt*, int iCol);
 const void *sqlite3_column_text16(sqlite3_stmt*, int iCol);
 sqlite3_value *sqlite3_column_value(sqlite3_stmt*, int iCol);
 int sqlite3_column_bytes(sqlite3_stmt*, int iCol);
 int sqlite3_column_bytes16(sqlite3_stmt*, int iCol);
 int sqlite3_column_type(sqlite3_stmt*, int iCol);
# 4998 "sqlite3.h"
 int sqlite3_finalize(sqlite3_stmt *pStmt);
# 5025 "sqlite3.h"
 int sqlite3_reset(sqlite3_stmt *pStmt);
# 5152 "sqlite3.h"
 int sqlite3_create_function(
  sqlite3 *db,
  const char *zFunctionName,
  int nArg,
  int eTextRep,
  void *pApp,
  void (*xFunc)(sqlite3_context*,int,sqlite3_value**),
  void (*xStep)(sqlite3_context*,int,sqlite3_value**),
  void (*xFinal)(sqlite3_context*)
);
 int sqlite3_create_function16(
  sqlite3 *db,
  const void *zFunctionName,
  int nArg,
  int eTextRep,
  void *pApp,
  void (*xFunc)(sqlite3_context*,int,sqlite3_value**),
  void (*xStep)(sqlite3_context*,int,sqlite3_value**),
  void (*xFinal)(sqlite3_context*)
);
 int sqlite3_create_function_v2(
  sqlite3 *db,
  const char *zFunctionName,
  int nArg,
  int eTextRep,
  void *pApp,
  void (*xFunc)(sqlite3_context*,int,sqlite3_value**),
  void (*xStep)(sqlite3_context*,int,sqlite3_value**),
  void (*xFinal)(sqlite3_context*),
  void(*xDestroy)(void*)
);
 int sqlite3_create_window_function(
  sqlite3 *db,
  const char *zFunctionName,
  int nArg,
  int eTextRep,
  void *pApp,
  void (*xStep)(sqlite3_context*,int,sqlite3_value**),
  void (*xFinal)(sqlite3_context*),
  void (*xValue)(sqlite3_context*),
  void (*xInverse)(sqlite3_context*,int,sqlite3_value**),
  void(*xDestroy)(void*)
);
# 5290 "sqlite3.h"
 int sqlite3_aggregate_count(sqlite3_context*);
 int sqlite3_expired(sqlite3_stmt*);
 int sqlite3_transfer_bindings(sqlite3_stmt*, sqlite3_stmt*);
 int sqlite3_global_recover(void);
 void sqlite3_thread_cleanup(void);
 int sqlite3_memory_alarm(void(*)(void*,sqlite3_int64,int),
                      void*,sqlite3_int64);
# 5427 "sqlite3.h"
 const void *sqlite3_value_blob(sqlite3_value*);
 double sqlite3_value_double(sqlite3_value*);
 int sqlite3_value_int(sqlite3_value*);
 sqlite3_int64 sqlite3_value_int64(sqlite3_value*);
 void *sqlite3_value_pointer(sqlite3_value*, const char*);
 const unsigned char *sqlite3_value_text(sqlite3_value*);
 const void *sqlite3_value_text16(sqlite3_value*);
 const void *sqlite3_value_text16le(sqlite3_value*);
 const void *sqlite3_value_text16be(sqlite3_value*);
 int sqlite3_value_bytes(sqlite3_value*);
 int sqlite3_value_bytes16(sqlite3_value*);
 int sqlite3_value_type(sqlite3_value*);
 int sqlite3_value_numeric_type(sqlite3_value*);
 int sqlite3_value_nochange(sqlite3_value*);
 int sqlite3_value_frombind(sqlite3_value*);
# 5453 "sqlite3.h"
 unsigned int sqlite3_value_subtype(sqlite3_value*);
# 5469 "sqlite3.h"
 sqlite3_value *sqlite3_value_dup(const sqlite3_value*);
 void sqlite3_value_free(sqlite3_value*);
# 5515 "sqlite3.h"
 void *sqlite3_aggregate_context(sqlite3_context*, int nBytes);
# 5530 "sqlite3.h"
 void *sqlite3_user_data(sqlite3_context*);
# 5542 "sqlite3.h"
 sqlite3 *sqlite3_context_db_handle(sqlite3_context*);
# 5601 "sqlite3.h"
 void *sqlite3_get_auxdata(sqlite3_context*, int N);
 void sqlite3_set_auxdata(sqlite3_context*, int N, void*, void (*)(void*));
# 5619 "sqlite3.h"
typedef void (*sqlite3_destructor_type)(void*);
# 5769 "sqlite3.h"
 void sqlite3_result_blob(sqlite3_context*, const void*, int, void(*)(void*));
 void sqlite3_result_blob64(sqlite3_context*,const void*,
                           sqlite3_uint64,void(*)(void*));
 void sqlite3_result_double(sqlite3_context*, double);
 void sqlite3_result_error(sqlite3_context*, const char*, int);
 void sqlite3_result_error16(sqlite3_context*, const void*, int);
 void sqlite3_result_error_toobig(sqlite3_context*);
 void sqlite3_result_error_nomem(sqlite3_context*);
 void sqlite3_result_error_code(sqlite3_context*, int);
 void sqlite3_result_int(sqlite3_context*, int);
 void sqlite3_result_int64(sqlite3_context*, sqlite3_int64);
 void sqlite3_result_null(sqlite3_context*);
 void sqlite3_result_text(sqlite3_context*, const char*, int, void(*)(void*));
 void sqlite3_result_text64(sqlite3_context*, const char*,sqlite3_uint64,
                           void(*)(void*), unsigned char encoding);
 void sqlite3_result_text16(sqlite3_context*, const void*, int, void(*)(void*));
 void sqlite3_result_text16le(sqlite3_context*, const void*, int,void(*)(void*));
 void sqlite3_result_text16be(sqlite3_context*, const void*, int,void(*)(void*));
 void sqlite3_result_value(sqlite3_context*, sqlite3_value*);
 void sqlite3_result_pointer(sqlite3_context*, void*,const char*,void(*)(void*));
 void sqlite3_result_zeroblob(sqlite3_context*, int n);
 int sqlite3_result_zeroblob64(sqlite3_context*, sqlite3_uint64 n);
# 5805 "sqlite3.h"
 void sqlite3_result_subtype(sqlite3_context*,unsigned int);
# 5888 "sqlite3.h"
 int sqlite3_create_collation(
  sqlite3*,
  const char *zName,
  int eTextRep,
  void *pArg,
  int(*xCompare)(void*,int,const void*,int,const void*)
);
 int sqlite3_create_collation_v2(
  sqlite3*,
  const char *zName,
  int eTextRep,
  void *pArg,
  int(*xCompare)(void*,int,const void*,int,const void*),
  void(*xDestroy)(void*)
);
 int sqlite3_create_collation16(
  sqlite3*,
  const void *zName,
  int eTextRep,
  void *pArg,
  int(*xCompare)(void*,int,const void*,int,const void*)
);
# 5938 "sqlite3.h"
 int sqlite3_collation_needed(
  sqlite3*,
  void*,
  void(*)(void*,sqlite3*,int eTextRep,const char*)
);
 int sqlite3_collation_needed16(
  sqlite3*,
  void*,
  void(*)(void*,sqlite3*,int eTextRep,const void*)
);
# 5976 "sqlite3.h"
 int sqlite3_sleep(int);
# 6034 "sqlite3.h"
 extern char *sqlite3_temp_directory;
# 6071 "sqlite3.h"
 extern char *sqlite3_data_directory;
# 6092 "sqlite3.h"
 int sqlite3_win32_set_directory(
  unsigned long type,
  void *zValue
);
 int sqlite3_win32_set_directory8(unsigned long type, const char *zValue);
 int sqlite3_win32_set_directory16(unsigned long type, const void *zValue);
# 6130 "sqlite3.h"
 int sqlite3_get_autocommit(sqlite3*);
# 6143 "sqlite3.h"
 sqlite3 *sqlite3_db_handle(sqlite3_stmt*);
# 6175 "sqlite3.h"
 const char *sqlite3_db_filename(sqlite3 *db, const char *zDbName);
# 6185 "sqlite3.h"
 int sqlite3_db_readonly(sqlite3 *db, const char *zDbName);
# 6201 "sqlite3.h"
 sqlite3_stmt *sqlite3_next_stmt(sqlite3 *pDb, sqlite3_stmt *pStmt);
# 6250 "sqlite3.h"
 void *sqlite3_commit_hook(sqlite3*, int(*)(void*), void*);
 void *sqlite3_rollback_hook(sqlite3*, void(*)(void *), void*);
# 6302 "sqlite3.h"
 void *sqlite3_update_hook(
  sqlite3*,
  void(*)(void *,int ,char const *,char const *,sqlite3_int64),
  void*
);
# 6347 "sqlite3.h"
 int sqlite3_enable_shared_cache(int);
# 6363 "sqlite3.h"
 int sqlite3_release_memory(int);
# 6377 "sqlite3.h"
 int sqlite3_db_release_memory(sqlite3*);
# 6443 "sqlite3.h"
 sqlite3_int64 sqlite3_soft_heap_limit64(sqlite3_int64 N);
 sqlite3_int64 sqlite3_hard_heap_limit64(sqlite3_int64 N);
# 6455 "sqlite3.h"
 void sqlite3_soft_heap_limit(int N);
# 6527 "sqlite3.h"
 int sqlite3_table_column_metadata(
  sqlite3 *db,
  const char *zDbName,
  const char *zTableName,
  const char *zColumnName,
  char const **pzDataType,
  char const **pzCollSeq,
  int *pNotNull,
  int *pPrimaryKey,
  int *pAutoinc
);
# 6583 "sqlite3.h"
 int sqlite3_load_extension(
  sqlite3 *db,
  const char *zFile,
  const char *zProc,
  char **pzErrMsg
);
# 6615 "sqlite3.h"
 int sqlite3_enable_load_extension(sqlite3 *db, int onoff);
# 6653 "sqlite3.h"
 int sqlite3_auto_extension(void(*xEntryPoint)(void));
# 6665 "sqlite3.h"
 int sqlite3_cancel_auto_extension(void(*xEntryPoint)(void));







 void sqlite3_reset_auto_extension(void);
# 6687 "sqlite3.h"
typedef struct sqlite3_vtab sqlite3_vtab;
typedef struct sqlite3_index_info sqlite3_index_info;
typedef struct sqlite3_vtab_cursor sqlite3_vtab_cursor;
typedef struct sqlite3_module sqlite3_module;
# 6708 "sqlite3.h"
struct sqlite3_module {
  int iVersion;
  int (*xCreate)(sqlite3*, void *pAux,
               int argc, const char *const*argv,
               sqlite3_vtab **ppVTab, char**);
  int (*xConnect)(sqlite3*, void *pAux,
               int argc, const char *const*argv,
               sqlite3_vtab **ppVTab, char**);
  int (*xBestIndex)(sqlite3_vtab *pVTab, sqlite3_index_info*);
  int (*xDisconnect)(sqlite3_vtab *pVTab);
  int (*xDestroy)(sqlite3_vtab *pVTab);
  int (*xOpen)(sqlite3_vtab *pVTab, sqlite3_vtab_cursor **ppCursor);
  int (*xClose)(sqlite3_vtab_cursor*);
  int (*xFilter)(sqlite3_vtab_cursor*, int idxNum, const char *idxStr,
                int argc, sqlite3_value **argv);
  int (*xNext)(sqlite3_vtab_cursor*);
  int (*xEof)(sqlite3_vtab_cursor*);
  int (*xColumn)(sqlite3_vtab_cursor*, sqlite3_context*, int);
  int (*xRowid)(sqlite3_vtab_cursor*, sqlite3_int64 *pRowid);
  int (*xUpdate)(sqlite3_vtab *, int, sqlite3_value **, sqlite3_int64 *);
  int (*xBegin)(sqlite3_vtab *pVTab);
  int (*xSync)(sqlite3_vtab *pVTab);
  int (*xCommit)(sqlite3_vtab *pVTab);
  int (*xRollback)(sqlite3_vtab *pVTab);
  int (*xFindFunction)(sqlite3_vtab *pVtab, int nArg, const char *zName,
                       void (**pxFunc)(sqlite3_context*,int,sqlite3_value**),
                       void **ppArg);
  int (*xRename)(sqlite3_vtab *pVtab, const char *zNew);


  int (*xSavepoint)(sqlite3_vtab *pVTab, int);
  int (*xRelease)(sqlite3_vtab *pVTab, int);
  int (*xRollbackTo)(sqlite3_vtab *pVTab, int);


  int (*xShadowName)(const char*);
};
# 6848 "sqlite3.h"
struct sqlite3_index_info {

  int nConstraint;
  struct sqlite3_index_constraint {
     int iColumn;
     unsigned char op;
     unsigned char usable;
     int iTermOffset;
  } *aConstraint;
  int nOrderBy;
  struct sqlite3_index_orderby {
     int iColumn;
     unsigned char desc;
  } *aOrderBy;

  struct sqlite3_index_constraint_usage {
    int argvIndex;
    unsigned char omit;
  } *aConstraintUsage;
  int idxNum;
  char *idxStr;
  int needToFreeIdxStr;
  int orderByConsumed;
  double estimatedCost;

  sqlite3_int64 estimatedRows;

  int idxFlags;

  sqlite3_uint64 colUsed;
};
# 6945 "sqlite3.h"
 int sqlite3_create_module(
  sqlite3 *db,
  const char *zName,
  const sqlite3_module *p,
  void *pClientData
);
 int sqlite3_create_module_v2(
  sqlite3 *db,
  const char *zName,
  const sqlite3_module *p,
  void *pClientData,
  void(*xDestroy)(void*)
);
# 6971 "sqlite3.h"
 int sqlite3_drop_modules(
  sqlite3 *db,
  const char **azKeep
);
# 6994 "sqlite3.h"
struct sqlite3_vtab {
  const sqlite3_module *pModule;
  int nRef;
  char *zErrMsg;

};
# 7018 "sqlite3.h"
struct sqlite3_vtab_cursor {
  sqlite3_vtab *pVtab;

};
# 7031 "sqlite3.h"
 int sqlite3_declare_vtab(sqlite3*, const char *zSQL);
# 7050 "sqlite3.h"
 int sqlite3_overload_function(sqlite3*, const char *zFuncName, int nArg);
# 7074 "sqlite3.h"
typedef struct sqlite3_blob sqlite3_blob;
# 7159 "sqlite3.h"
 int sqlite3_blob_open(
  sqlite3*,
  const char *zDb,
  const char *zTable,
  const char *zColumn,
  sqlite3_int64 iRow,
  int flags,
  sqlite3_blob **ppBlob
);
# 7192 "sqlite3.h"
 int sqlite3_blob_reopen(sqlite3_blob *, sqlite3_int64);
# 7215 "sqlite3.h"
 int sqlite3_blob_close(sqlite3_blob *);
# 7231 "sqlite3.h"
 int sqlite3_blob_bytes(sqlite3_blob *);
# 7260 "sqlite3.h"
 int sqlite3_blob_read(sqlite3_blob *, void *Z, int N, int iOffset);
# 7302 "sqlite3.h"
 int sqlite3_blob_write(sqlite3_blob *, const void *z, int n, int iOffset);
# 7333 "sqlite3.h"
 sqlite3_vfs *sqlite3_vfs_find(const char *zVfsName);
 int sqlite3_vfs_register(sqlite3_vfs*, int makeDflt);
 int sqlite3_vfs_unregister(sqlite3_vfs*);
# 7451 "sqlite3.h"
 sqlite3_mutex *sqlite3_mutex_alloc(int);
 void sqlite3_mutex_free(sqlite3_mutex*);
 void sqlite3_mutex_enter(sqlite3_mutex*);
 int sqlite3_mutex_try(sqlite3_mutex*);
 void sqlite3_mutex_leave(sqlite3_mutex*);
# 7522 "sqlite3.h"
typedef struct sqlite3_mutex_methods sqlite3_mutex_methods;
struct sqlite3_mutex_methods {
  int (*xMutexInit)(void);
  int (*xMutexEnd)(void);
  sqlite3_mutex *(*xMutexAlloc)(int);
  void (*xMutexFree)(sqlite3_mutex *);
  void (*xMutexEnter)(sqlite3_mutex *);
  int (*xMutexTry)(sqlite3_mutex *);
  void (*xMutexLeave)(sqlite3_mutex *);
  int (*xMutexHeld)(sqlite3_mutex *);
  int (*xMutexNotheld)(sqlite3_mutex *);
};
# 7565 "sqlite3.h"
 int sqlite3_mutex_held(sqlite3_mutex*);
 int sqlite3_mutex_notheld(sqlite3_mutex*);
# 7606 "sqlite3.h"
 sqlite3_mutex *sqlite3_db_mutex(sqlite3*);
# 7649 "sqlite3.h"
 int sqlite3_file_control(sqlite3*, const char *zDbName, int op, void*);
# 7668 "sqlite3.h"
 int sqlite3_test_control(int op, ...);
# 7758 "sqlite3.h"
 int sqlite3_keyword_count(void);
 int sqlite3_keyword_name(int,const char**,int*);
 int sqlite3_keyword_check(const char*,int);
# 7778 "sqlite3.h"
typedef struct sqlite3_str sqlite3_str;
# 7805 "sqlite3.h"
 sqlite3_str *sqlite3_str_new(sqlite3*);
# 7820 "sqlite3.h"
 char *sqlite3_str_finish(sqlite3_str*);
# 7854 "sqlite3.h"
 void sqlite3_str_appendf(sqlite3_str*, const char *zFormat, ...);
 void sqlite3_str_vappendf(sqlite3_str*, const char *zFormat, va_list);
 void sqlite3_str_append(sqlite3_str*, const char *zIn, int N);
 void sqlite3_str_appendall(sqlite3_str*, const char *zIn);
 void sqlite3_str_appendchar(sqlite3_str*, int N, char C);
 void sqlite3_str_reset(sqlite3_str*);
# 7890 "sqlite3.h"
 int sqlite3_str_errcode(sqlite3_str*);
 int sqlite3_str_length(sqlite3_str*);
 char *sqlite3_str_value(sqlite3_str*);
# 7920 "sqlite3.h"
 int sqlite3_status(int op, int *pCurrent, int *pHighwater, int resetFlag);
 int sqlite3_status64(
  int op,
  sqlite3_int64 *pCurrent,
  sqlite3_int64 *pHighwater,
  int resetFlag
);
# 8030 "sqlite3.h"
 int sqlite3_db_status(sqlite3*, int op, int *pCur, int *pHiwtr, int resetFlg);
# 8183 "sqlite3.h"
 int sqlite3_stmt_status(sqlite3_stmt*, int op,int resetFlg);
# 8259 "sqlite3.h"
typedef struct sqlite3_pcache sqlite3_pcache;
# 8271 "sqlite3.h"
typedef struct sqlite3_pcache_page sqlite3_pcache_page;
struct sqlite3_pcache_page {
  void *pBuf;
  void *pExtra;
};
# 8436 "sqlite3.h"
typedef struct sqlite3_pcache_methods2 sqlite3_pcache_methods2;
struct sqlite3_pcache_methods2 {
  int iVersion;
  void *pArg;
  int (*xInit)(void*);
  void (*xShutdown)(void*);
  sqlite3_pcache *(*xCreate)(int szPage, int szExtra, int bPurgeable);
  void (*xCachesize)(sqlite3_pcache*, int nCachesize);
  int (*xPagecount)(sqlite3_pcache*);
  sqlite3_pcache_page *(*xFetch)(sqlite3_pcache*, unsigned key, int createFlag);
  void (*xUnpin)(sqlite3_pcache*, sqlite3_pcache_page*, int discard);
  void (*xRekey)(sqlite3_pcache*, sqlite3_pcache_page*,
      unsigned oldKey, unsigned newKey);
  void (*xTruncate)(sqlite3_pcache*, unsigned iLimit);
  void (*xDestroy)(sqlite3_pcache*);
  void (*xShrink)(sqlite3_pcache*);
};






typedef struct sqlite3_pcache_methods sqlite3_pcache_methods;
struct sqlite3_pcache_methods {
  void *pArg;
  int (*xInit)(void*);
  void (*xShutdown)(void*);
  sqlite3_pcache *(*xCreate)(int szPage, int bPurgeable);
  void (*xCachesize)(sqlite3_pcache*, int nCachesize);
  int (*xPagecount)(sqlite3_pcache*);
  void *(*xFetch)(sqlite3_pcache*, unsigned key, int createFlag);
  void (*xUnpin)(sqlite3_pcache*, void*, int discard);
  void (*xRekey)(sqlite3_pcache*, void*, unsigned oldKey, unsigned newKey);
  void (*xTruncate)(sqlite3_pcache*, unsigned iLimit);
  void (*xDestroy)(sqlite3_pcache*);
};
# 8485 "sqlite3.h"
typedef struct sqlite3_backup sqlite3_backup;
# 8673 "sqlite3.h"
 sqlite3_backup *sqlite3_backup_init(
  sqlite3 *pDest,
  const char *zDestName,
  sqlite3 *pSource,
  const char *zSourceName
);
 int sqlite3_backup_step(sqlite3_backup *p, int nPage);
 int sqlite3_backup_finish(sqlite3_backup *p);
 int sqlite3_backup_remaining(sqlite3_backup *p);
 int sqlite3_backup_pagecount(sqlite3_backup *p);
# 8799 "sqlite3.h"
 int sqlite3_unlock_notify(
  sqlite3 *pBlocked,
  void (*xNotify)(void **apArg, int nArg),
  void *pNotifyArg
);
# 8814 "sqlite3.h"
 int sqlite3_stricmp(const char *, const char *);
 int sqlite3_strnicmp(const char *, const char *, int);
# 8832 "sqlite3.h"
 int sqlite3_strglob(const char *zGlob, const char *zStr);
# 8855 "sqlite3.h"
 int sqlite3_strlike(const char *zGlob, const char *zStr, unsigned int cEsc);
# 8878 "sqlite3.h"
 void sqlite3_log(int iErrCode, const char *zFormat, ...);
# 8914 "sqlite3.h"
 void *sqlite3_wal_hook(
  sqlite3*,
  int(*)(void *,sqlite3*,const char*,int),
  void*
);
# 8949 "sqlite3.h"
 int sqlite3_wal_autocheckpoint(sqlite3 *db, int N);
# 8971 "sqlite3.h"
 int sqlite3_wal_checkpoint(sqlite3 *db, const char *zDb);
# 9065 "sqlite3.h"
 int sqlite3_wal_checkpoint_v2(
  sqlite3 *db,
  const char *zDb,
  int eMode,
  int *pnLog,
  int *pnCkpt
);
# 9105 "sqlite3.h"
 int sqlite3_vtab_config(sqlite3*, int op, ...);
# 9183 "sqlite3.h"
 int sqlite3_vtab_on_conflict(sqlite3 *);
# 9202 "sqlite3.h"
 int sqlite3_vtab_nochange(sqlite3_context*);
# 9217 "sqlite3.h"
 const char *sqlite3_vtab_collation(sqlite3_index_info*,int);
# 9322 "sqlite3.h"
 int sqlite3_stmt_scanstatus(
  sqlite3_stmt *pStmt,
  int idx,
  int iScanStatusOp,
  void *pOut
);
# 9338 "sqlite3.h"
 void sqlite3_stmt_scanstatus_reset(sqlite3_stmt*);
# 9370 "sqlite3.h"
 int sqlite3_db_cacheflush(sqlite3*);
# 9484 "sqlite3.h"
 int sqlite3_system_errno(sqlite3*);
# 9506 "sqlite3.h"
typedef struct sqlite3_snapshot {
  unsigned char hidden[48];
} sqlite3_snapshot;
# 9553 "sqlite3.h"
 int sqlite3_snapshot_get(
  sqlite3 *db,
  const char *zSchema,
  sqlite3_snapshot **ppSnapshot
);
# 9602 "sqlite3.h"
 int sqlite3_snapshot_open(
  sqlite3 *db,
  const char *zSchema,
  sqlite3_snapshot *pSnapshot
);
# 9619 "sqlite3.h"
 void sqlite3_snapshot_free(sqlite3_snapshot*);
# 9646 "sqlite3.h"
 int sqlite3_snapshot_cmp(
  sqlite3_snapshot *p1,
  sqlite3_snapshot *p2
);
# 9674 "sqlite3.h"
 int sqlite3_snapshot_recover(sqlite3 *db, const char *zDb);
# 9712 "sqlite3.h"
 unsigned char *sqlite3_serialize(
  sqlite3 *db,
  const char *zSchema,
  sqlite3_int64 *piSize,
  unsigned int mFlags
);
# 9764 "sqlite3.h"
 int sqlite3_deserialize(
  sqlite3 *db,
  const char *zSchema,
  unsigned char *pData,
  sqlite3_int64 szDb,
  sqlite3_int64 szBuf,
  unsigned mFlags
);
# 9807 "sqlite3.h"
}
# 9830 "sqlite3.h"
extern "C" {


typedef struct sqlite3_rtree_geometry sqlite3_rtree_geometry;
typedef struct sqlite3_rtree_query_info sqlite3_rtree_query_info;







  typedef double sqlite3_rtree_dbl;
# 9851 "sqlite3.h"
 int sqlite3_rtree_geometry_callback(
  sqlite3 *db,
  const char *zGeom,
  int (*xGeom)(sqlite3_rtree_geometry*, int, sqlite3_rtree_dbl*,int*),
  void *pContext
);






struct sqlite3_rtree_geometry {
  void *pContext;
  int nParam;
  sqlite3_rtree_dbl *aParam;
  void *pUser;
  void (*xDelUser)(void *);
};







 int sqlite3_rtree_query_callback(
  sqlite3 *db,
  const char *zQueryFunc,
  int (*xQueryFunc)(sqlite3_rtree_query_info*),
  void *pContext,
  void (*xDestructor)(void*)
);
# 9895 "sqlite3.h"
struct sqlite3_rtree_query_info {
  void *pContext;
  int nParam;
  sqlite3_rtree_dbl *aParam;
  void *pUser;
  void (*xDelUser)(void*);
  sqlite3_rtree_dbl *aCoord;
  unsigned int *anQueue;
  int nCoord;
  int iLevel;
  int mxLevel;
  sqlite3_int64 iRowid;
  sqlite3_rtree_dbl rParentScore;
  int eParentWithin;
  int eWithin;
  sqlite3_rtree_dbl rScore;

  sqlite3_value **apSqlParam;
};
# 9924 "sqlite3.h"
}
# 11618 "sqlite3.h"
extern "C" {
# 11628 "sqlite3.h"
typedef struct Fts5ExtensionApi Fts5ExtensionApi;
typedef struct Fts5Context Fts5Context;
typedef struct Fts5PhraseIter Fts5PhraseIter;

typedef void (*fts5_extension_function)(
  const Fts5ExtensionApi *pApi,
  Fts5Context *pFts,
  sqlite3_context *pCtx,
  int nVal,
  sqlite3_value **apVal
);

struct Fts5PhraseIter {
  const unsigned char *a;
  const unsigned char *b;
};
# 11856 "sqlite3.h"
struct Fts5ExtensionApi {
  int iVersion;

  void *(*xUserData)(Fts5Context*);

  int (*xColumnCount)(Fts5Context*);
  int (*xRowCount)(Fts5Context*, sqlite3_int64 *pnRow);
  int (*xColumnTotalSize)(Fts5Context*, int iCol, sqlite3_int64 *pnToken);

  int (*xTokenize)(Fts5Context*,
    const char *pText, int nText,
    void *pCtx,
    int (*xToken)(void*, int, const char*, int, int, int)
  );

  int (*xPhraseCount)(Fts5Context*);
  int (*xPhraseSize)(Fts5Context*, int iPhrase);

  int (*xInstCount)(Fts5Context*, int *pnInst);
  int (*xInst)(Fts5Context*, int iIdx, int *piPhrase, int *piCol, int *piOff);

  sqlite3_int64 (*xRowid)(Fts5Context*);
  int (*xColumnText)(Fts5Context*, int iCol, const char **pz, int *pn);
  int (*xColumnSize)(Fts5Context*, int iCol, int *pnToken);

  int (*xQueryPhrase)(Fts5Context*, int iPhrase, void *pUserData,
    int(*)(const Fts5ExtensionApi*,Fts5Context*,void*)
  );
  int (*xSetAuxdata)(Fts5Context*, void *pAux, void(*xDelete)(void*));
  void *(*xGetAuxdata)(Fts5Context*, int bClear);

  int (*xPhraseFirst)(Fts5Context*, int iPhrase, Fts5PhraseIter*, int*, int*);
  void (*xPhraseNext)(Fts5Context*, Fts5PhraseIter*, int *piCol, int *piOff);

  int (*xPhraseFirstColumn)(Fts5Context*, int iPhrase, Fts5PhraseIter*, int*);
  void (*xPhraseNextColumn)(Fts5Context*, Fts5PhraseIter*, int *piCol);
};
# 12090 "sqlite3.h"
typedef struct Fts5Tokenizer Fts5Tokenizer;
typedef struct fts5_tokenizer fts5_tokenizer;
struct fts5_tokenizer {
  int (*xCreate)(void*, const char **azArg, int nArg, Fts5Tokenizer **ppOut);
  void (*xDelete)(Fts5Tokenizer*);
  int (*xTokenize)(Fts5Tokenizer*,
      void *pCtx,
      int flags,
      const char *pText, int nText,
      int (*xToken)(
        void *pCtx,
        int tflags,
        const char *pToken,
        int nToken,
        int iStart,
        int iEnd
      )
  );
};
# 12127 "sqlite3.h"
typedef struct fts5_api fts5_api;
struct fts5_api {
  int iVersion;


  int (*xCreateTokenizer)(
    fts5_api *pApi,
    const char *zName,
    void *pContext,
    fts5_tokenizer *pTokenizer,
    void (*xDestroy)(void*)
  );


  int (*xFindTokenizer)(
    fts5_api *pApi,
    const char *zName,
    void **ppContext,
    fts5_tokenizer *pTokenizer
  );


  int (*xCreateFunction)(
    fts5_api *pApi,
    const char *zName,
    void *pContext,
    fts5_extension_function xFunction,
    void (*xDestroy)(void*)
  );
};






}
# 21 "sqlite3ext.h" 2
# 32 "sqlite3ext.h"
struct sqlite3_api_routines {
  void * (*aggregate_context)(sqlite3_context*,int nBytes);
  int (*aggregate_count)(sqlite3_context*);
  int (*bind_blob)(sqlite3_stmt*,int,const void*,int n,void(*)(void*));
  int (*bind_double)(sqlite3_stmt*,int,double);
  int (*bind_int)(sqlite3_stmt*,int,int);
  int (*bind_int64)(sqlite3_stmt*,int,sqlite_int64);
  int (*bind_null)(sqlite3_stmt*,int);
  int (*bind_parameter_count)(sqlite3_stmt*);
  int (*bind_parameter_index)(sqlite3_stmt*,const char*zName);
  const char * (*bind_parameter_name)(sqlite3_stmt*,int);
  int (*bind_text)(sqlite3_stmt*,int,const char*,int n,void(*)(void*));
  int (*bind_text16)(sqlite3_stmt*,int,const void*,int,void(*)(void*));
  int (*bind_value)(sqlite3_stmt*,int,const sqlite3_value*);
  int (*busy_handler)(sqlite3*,int(*)(void*,int),void*);
  int (*busy_timeout)(sqlite3*,int ms);
  int (*changes)(sqlite3*);
  int (*close)(sqlite3*);
  int (*collation_needed)(sqlite3*,void*,void(*)(void*,sqlite3*,
                           int eTextRep,const char*));
  int (*collation_needed16)(sqlite3*,void*,void(*)(void*,sqlite3*,
                             int eTextRep,const void*));
  const void * (*column_blob)(sqlite3_stmt*,int iCol);
  int (*column_bytes)(sqlite3_stmt*,int iCol);
  int (*column_bytes16)(sqlite3_stmt*,int iCol);
  int (*column_count)(sqlite3_stmt*pStmt);
  const char * (*column_database_name)(sqlite3_stmt*,int);
  const void * (*column_database_name16)(sqlite3_stmt*,int);
  const char * (*column_decltype)(sqlite3_stmt*,int i);
  const void * (*column_decltype16)(sqlite3_stmt*,int);
  double (*column_double)(sqlite3_stmt*,int iCol);
  int (*column_int)(sqlite3_stmt*,int iCol);
  sqlite_int64 (*column_int64)(sqlite3_stmt*,int iCol);
  const char * (*column_name)(sqlite3_stmt*,int);
  const void * (*column_name16)(sqlite3_stmt*,int);
  const char * (*column_origin_name)(sqlite3_stmt*,int);
  const void * (*column_origin_name16)(sqlite3_stmt*,int);
  const char * (*column_table_name)(sqlite3_stmt*,int);
  const void * (*column_table_name16)(sqlite3_stmt*,int);
  const unsigned char * (*column_text)(sqlite3_stmt*,int iCol);
  const void * (*column_text16)(sqlite3_stmt*,int iCol);
  int (*column_type)(sqlite3_stmt*,int iCol);
  sqlite3_value* (*column_value)(sqlite3_stmt*,int iCol);
  void * (*commit_hook)(sqlite3*,int(*)(void*),void*);
  int (*complete)(const char*sql);
  int (*complete16)(const void*sql);
  int (*create_collation)(sqlite3*,const char*,int,void*,
                           int(*)(void*,int,const void*,int,const void*));
  int (*create_collation16)(sqlite3*,const void*,int,void*,
                             int(*)(void*,int,const void*,int,const void*));
  int (*create_function)(sqlite3*,const char*,int,int,void*,
                          void (*xFunc)(sqlite3_context*,int,sqlite3_value**),
                          void (*xStep)(sqlite3_context*,int,sqlite3_value**),
                          void (*xFinal)(sqlite3_context*));
  int (*create_function16)(sqlite3*,const void*,int,int,void*,
                            void (*xFunc)(sqlite3_context*,int,sqlite3_value**),
                            void (*xStep)(sqlite3_context*,int,sqlite3_value**),
                            void (*xFinal)(sqlite3_context*));
  int (*create_module)(sqlite3*,const char*,const sqlite3_module*,void*);
  int (*data_count)(sqlite3_stmt*pStmt);
  sqlite3 * (*db_handle)(sqlite3_stmt*);
  int (*declare_vtab)(sqlite3*,const char*);
  int (*enable_shared_cache)(int);
  int (*errcode)(sqlite3*db);
  const char * (*errmsg)(sqlite3*);
  const void * (*errmsg16)(sqlite3*);
  int (*exec)(sqlite3*,const char*,sqlite3_callback,void*,char**);
  int (*expired)(sqlite3_stmt*);
  int (*finalize)(sqlite3_stmt*pStmt);
  void (*free)(void*);
  void (*free_table)(char**result);
  int (*get_autocommit)(sqlite3*);
  void * (*get_auxdata)(sqlite3_context*,int);
  int (*get_table)(sqlite3*,const char*,char***,int*,int*,char**);
  int (*global_recover)(void);
  void (*interruptx)(sqlite3*);
  sqlite_int64 (*last_insert_rowid)(sqlite3*);
  const char * (*libversion)(void);
  int (*libversion_number)(void);
  void *(*malloc)(int);
  char * (*mprintf)(const char*,...);
  int (*open)(const char*,sqlite3**);
  int (*open16)(const void*,sqlite3**);
  int (*prepare)(sqlite3*,const char*,int,sqlite3_stmt**,const char**);
  int (*prepare16)(sqlite3*,const void*,int,sqlite3_stmt**,const void**);
  void * (*profile)(sqlite3*,void(*)(void*,const char*,sqlite_uint64),void*);
  void (*progress_handler)(sqlite3*,int,int(*)(void*),void*);
  void *(*realloc)(void*,int);
  int (*reset)(sqlite3_stmt*pStmt);
  void (*result_blob)(sqlite3_context*,const void*,int,void(*)(void*));
  void (*result_double)(sqlite3_context*,double);
  void (*result_error)(sqlite3_context*,const char*,int);
  void (*result_error16)(sqlite3_context*,const void*,int);
  void (*result_int)(sqlite3_context*,int);
  void (*result_int64)(sqlite3_context*,sqlite_int64);
  void (*result_null)(sqlite3_context*);
  void (*result_text)(sqlite3_context*,const char*,int,void(*)(void*));
  void (*result_text16)(sqlite3_context*,const void*,int,void(*)(void*));
  void (*result_text16be)(sqlite3_context*,const void*,int,void(*)(void*));
  void (*result_text16le)(sqlite3_context*,const void*,int,void(*)(void*));
  void (*result_value)(sqlite3_context*,sqlite3_value*);
  void * (*rollback_hook)(sqlite3*,void(*)(void*),void*);
  int (*set_authorizer)(sqlite3*,int(*)(void*,int,const char*,const char*,
                         const char*,const char*),void*);
  void (*set_auxdata)(sqlite3_context*,int,void*,void (*)(void*));
  char * (*xsnprintf)(int,char*,const char*,...);
  int (*step)(sqlite3_stmt*);
  int (*table_column_metadata)(sqlite3*,const char*,const char*,const char*,
                                char const**,char const**,int*,int*,int*);
  void (*thread_cleanup)(void);
  int (*total_changes)(sqlite3*);
  void * (*trace)(sqlite3*,void(*xTrace)(void*,const char*),void*);
  int (*transfer_bindings)(sqlite3_stmt*,sqlite3_stmt*);
  void * (*update_hook)(sqlite3*,void(*)(void*,int ,char const*,char const*,
                                         sqlite_int64),void*);
  void * (*user_data)(sqlite3_context*);
  const void * (*value_blob)(sqlite3_value*);
  int (*value_bytes)(sqlite3_value*);
  int (*value_bytes16)(sqlite3_value*);
  double (*value_double)(sqlite3_value*);
  int (*value_int)(sqlite3_value*);
  sqlite_int64 (*value_int64)(sqlite3_value*);
  int (*value_numeric_type)(sqlite3_value*);
  const unsigned char * (*value_text)(sqlite3_value*);
  const void * (*value_text16)(sqlite3_value*);
  const void * (*value_text16be)(sqlite3_value*);
  const void * (*value_text16le)(sqlite3_value*);
  int (*value_type)(sqlite3_value*);
  char *(*vmprintf)(const char*,va_list);

  int (*overload_function)(sqlite3*, const char *zFuncName, int nArg);

  int (*prepare_v2)(sqlite3*,const char*,int,sqlite3_stmt**,const char**);
  int (*prepare16_v2)(sqlite3*,const void*,int,sqlite3_stmt**,const void**);
  int (*clear_bindings)(sqlite3_stmt*);

  int (*create_module_v2)(sqlite3*,const char*,const sqlite3_module*,void*,
                          void (*xDestroy)(void *));

  int (*bind_zeroblob)(sqlite3_stmt*,int,int);
  int (*blob_bytes)(sqlite3_blob*);
  int (*blob_close)(sqlite3_blob*);
  int (*blob_open)(sqlite3*,const char*,const char*,const char*,sqlite3_int64,
                   int,sqlite3_blob**);
  int (*blob_read)(sqlite3_blob*,void*,int,int);
  int (*blob_write)(sqlite3_blob*,const void*,int,int);
  int (*create_collation_v2)(sqlite3*,const char*,int,void*,
                             int(*)(void*,int,const void*,int,const void*),
                             void(*)(void*));
  int (*file_control)(sqlite3*,const char*,int,void*);
  sqlite3_int64 (*memory_highwater)(int);
  sqlite3_int64 (*memory_used)(void);
  sqlite3_mutex *(*mutex_alloc)(int);
  void (*mutex_enter)(sqlite3_mutex*);
  void (*mutex_free)(sqlite3_mutex*);
  void (*mutex_leave)(sqlite3_mutex*);
  int (*mutex_try)(sqlite3_mutex*);
  int (*open_v2)(const char*,sqlite3**,int,const char*);
  int (*release_memory)(int);
  void (*result_error_nomem)(sqlite3_context*);
  void (*result_error_toobig)(sqlite3_context*);
  int (*sleep)(int);
  void (*soft_heap_limit)(int);
  sqlite3_vfs *(*vfs_find)(const char*);
  int (*vfs_register)(sqlite3_vfs*,int);
  int (*vfs_unregister)(sqlite3_vfs*);
  int (*xthreadsafe)(void);
  void (*result_zeroblob)(sqlite3_context*,int);
  void (*result_error_code)(sqlite3_context*,int);
  int (*test_control)(int, ...);
  void (*randomness)(int,void*);
  sqlite3 *(*context_db_handle)(sqlite3_context*);
  int (*extended_result_codes)(sqlite3*,int);
  int (*limit)(sqlite3*,int,int);
  sqlite3_stmt *(*next_stmt)(sqlite3*,sqlite3_stmt*);
  const char *(*sql)(sqlite3_stmt*);
  int (*status)(int,int*,int*,int);
  int (*backup_finish)(sqlite3_backup*);
  sqlite3_backup *(*backup_init)(sqlite3*,const char*,sqlite3*,const char*);
  int (*backup_pagecount)(sqlite3_backup*);
  int (*backup_remaining)(sqlite3_backup*);
  int (*backup_step)(sqlite3_backup*,int);
  const char *(*compileoption_get)(int);
  int (*compileoption_used)(const char*);
  int (*create_function_v2)(sqlite3*,const char*,int,int,void*,
                            void (*xFunc)(sqlite3_context*,int,sqlite3_value**),
                            void (*xStep)(sqlite3_context*,int,sqlite3_value**),
                            void (*xFinal)(sqlite3_context*),
                            void(*xDestroy)(void*));
  int (*db_config)(sqlite3*,int,...);
  sqlite3_mutex *(*db_mutex)(sqlite3*);
  int (*db_status)(sqlite3*,int,int*,int*,int);
  int (*extended_errcode)(sqlite3*);
  void (*log)(int,const char*,...);
  sqlite3_int64 (*soft_heap_limit64)(sqlite3_int64);
  const char *(*sourceid)(void);
  int (*stmt_status)(sqlite3_stmt*,int,int);
  int (*strnicmp)(const char*,const char*,int);
  int (*unlock_notify)(sqlite3*,void(*)(void**,int),void*);
  int (*wal_autocheckpoint)(sqlite3*,int);
  int (*wal_checkpoint)(sqlite3*,const char*);
  void *(*wal_hook)(sqlite3*,int(*)(void*,sqlite3*,const char*,int),void*);
  int (*blob_reopen)(sqlite3_blob*,sqlite3_int64);
  int (*vtab_config)(sqlite3*,int op,...);
  int (*vtab_on_conflict)(sqlite3*);

  int (*close_v2)(sqlite3*);
  const char *(*db_filename)(sqlite3*,const char*);
  int (*db_readonly)(sqlite3*,const char*);
  int (*db_release_memory)(sqlite3*);
  const char *(*errstr)(int);
  int (*stmt_busy)(sqlite3_stmt*);
  int (*stmt_readonly)(sqlite3_stmt*);
  int (*stricmp)(const char*,const char*);
  int (*uri_boolean)(const char*,const char*,int);
  sqlite3_int64 (*uri_int64)(const char*,const char*,sqlite3_int64);
  const char *(*uri_parameter)(const char*,const char*);
  char *(*xvsnprintf)(int,char*,const char*,va_list);
  int (*wal_checkpoint_v2)(sqlite3*,const char*,int,int*,int*);

  int (*auto_extension)(void(*)(void));
  int (*bind_blob64)(sqlite3_stmt*,int,const void*,sqlite3_uint64,
                     void(*)(void*));
  int (*bind_text64)(sqlite3_stmt*,int,const char*,sqlite3_uint64,
                      void(*)(void*),unsigned char);
  int (*cancel_auto_extension)(void(*)(void));
  int (*load_extension)(sqlite3*,const char*,const char*,char**);
  void *(*malloc64)(sqlite3_uint64);
  sqlite3_uint64 (*msize)(void*);
  void *(*realloc64)(void*,sqlite3_uint64);
  void (*reset_auto_extension)(void);
  void (*result_blob64)(sqlite3_context*,const void*,sqlite3_uint64,
                        void(*)(void*));
  void (*result_text64)(sqlite3_context*,const char*,sqlite3_uint64,
                         void(*)(void*), unsigned char);
  int (*strglob)(const char*,const char*);

  sqlite3_value *(*value_dup)(const sqlite3_value*);
  void (*value_free)(sqlite3_value*);
  int (*result_zeroblob64)(sqlite3_context*,sqlite3_uint64);
  int (*bind_zeroblob64)(sqlite3_stmt*, int, sqlite3_uint64);

  unsigned int (*value_subtype)(sqlite3_value*);
  void (*result_subtype)(sqlite3_context*,unsigned int);

  int (*status64)(int,sqlite3_int64*,sqlite3_int64*,int);
  int (*strlike)(const char*,const char*,unsigned int);
  int (*db_cacheflush)(sqlite3*);

  int (*system_errno)(sqlite3*);

  int (*trace_v2)(sqlite3*,unsigned,int(*)(unsigned,void*,void*,void*),void*);
  char *(*expanded_sql)(sqlite3_stmt*);

  void (*set_last_insert_rowid)(sqlite3*,sqlite3_int64);

  int (*prepare_v3)(sqlite3*,const char*,int,unsigned int,
                    sqlite3_stmt**,const char**);
  int (*prepare16_v3)(sqlite3*,const void*,int,unsigned int,
                      sqlite3_stmt**,const void**);
  int (*bind_pointer)(sqlite3_stmt*,int,void*,const char*,void(*)(void*));
  void (*result_pointer)(sqlite3_context*,void*,const char*,void(*)(void*));
  void *(*value_pointer)(sqlite3_value*,const char*);
  int (*vtab_nochange)(sqlite3_context*);
  int (*value_nochange)(sqlite3_value*);
  const char *(*vtab_collation)(sqlite3_index_info*,int);

  int (*keyword_count)(void);
  int (*keyword_name)(int,const char**,int*);
  int (*keyword_check)(const char*,int);
  sqlite3_str *(*str_new)(sqlite3*);
  char *(*str_finish)(sqlite3_str*);
  void (*str_appendf)(sqlite3_str*, const char *zFormat, ...);
  void (*str_vappendf)(sqlite3_str*, const char *zFormat, va_list);
  void (*str_append)(sqlite3_str*, const char *zIn, int N);
  void (*str_appendall)(sqlite3_str*, const char *zIn);
  void (*str_appendchar)(sqlite3_str*, int N, char C);
  void (*str_reset)(sqlite3_str*);
  int (*str_errcode)(sqlite3_str*);
  int (*str_length)(sqlite3_str*);
  char *(*str_value)(sqlite3_str*);

  int (*create_window_function)(sqlite3*,const char*,int,int,void*,
                            void (*xStep)(sqlite3_context*,int,sqlite3_value**),
                            void (*xFinal)(sqlite3_context*),
                            void (*xValue)(sqlite3_context*),
                            void (*xInv)(sqlite3_context*,int,sqlite3_value**),
                            void(*xDestroy)(void*));

  const char *(*normalized_sql)(sqlite3_stmt*);

  int (*stmt_isexplain)(sqlite3_stmt*);
  int (*value_frombind)(sqlite3_value*);

  int (*drop_modules)(sqlite3*,const char**);

  sqlite3_int64 (*hard_heap_limit64)(sqlite3_int64);
  const char *(*uri_key)(const char*,int);
  const char *(*filename_database)(const char*);
  const char *(*filename_journal)(const char*);
  const char *(*filename_wal)(const char*);

  char *(*create_filename)(const char*,const char*,const char*,
                           int,const char**);
  void (*free_filename)(char*);
  sqlite3_file *(*database_file_object)(const char*);
};





typedef int (*sqlite3_loadext_entry)(
  sqlite3 *db,
  char **pzErrMsg,
  const sqlite3_api_routines *pThunk
);
# 29 "RegistExt.h" 2

# 1 "/usr/include/c++/10/cstddef" 1 3
# 42 "/usr/include/c++/10/cstddef" 3
       
# 43 "/usr/include/c++/10/cstddef" 3






# 1 "/usr/include/c++/10/x86_64-redhat-linux/bits/c++config.h" 1 3


# 1 "/usr/include/bits/wordsize.h" 1 3 4
# 4 "/usr/include/c++/10/x86_64-redhat-linux/bits/c++config.h" 2 3
# 2353 "/usr/include/c++/10/x86_64-redhat-linux/bits/c++config.h" 3

# 2353 "/usr/include/c++/10/x86_64-redhat-linux/bits/c++config.h" 3
namespace std
{
  typedef long unsigned int size_t;
  typedef long int ptrdiff_t;


  typedef decltype(nullptr) nullptr_t;

}
# 2375 "/usr/include/c++/10/x86_64-redhat-linux/bits/c++config.h" 3
namespace std
{
  inline namespace __cxx11 __attribute__((__abi_tag__ ("cxx11"))) { }
}
namespace __gnu_cxx
{
  inline namespace __cxx11 __attribute__((__abi_tag__ ("cxx11"))) { }
}
# 2613 "/usr/include/c++/10/x86_64-redhat-linux/bits/c++config.h" 3
# 1 "/usr/include/c++/10/x86_64-redhat-linux/bits/os_defines.h" 1 3
# 39 "/usr/include/c++/10/x86_64-redhat-linux/bits/os_defines.h" 3
# 1 "/usr/include/features.h" 1 3 4
# 465 "/usr/include/features.h" 3 4
# 1 "/usr/include/sys/cdefs.h" 1 3 4
# 452 "/usr/include/sys/cdefs.h" 3 4
# 1 "/usr/include/bits/wordsize.h" 1 3 4
# 453 "/usr/include/sys/cdefs.h" 2 3 4
# 1 "/usr/include/bits/long-double.h" 1 3 4
# 454 "/usr/include/sys/cdefs.h" 2 3 4
# 466 "/usr/include/features.h" 2 3 4
# 489 "/usr/include/features.h" 3 4
# 1 "/usr/include/gnu/stubs.h" 1 3 4
# 10 "/usr/include/gnu/stubs.h" 3 4
# 1 "/usr/include/gnu/stubs-64.h" 1 3 4
# 11 "/usr/include/gnu/stubs.h" 2 3 4
# 490 "/usr/include/features.h" 2 3 4
# 40 "/usr/include/c++/10/x86_64-redhat-linux/bits/os_defines.h" 2 3
# 2614 "/usr/include/c++/10/x86_64-redhat-linux/bits/c++config.h" 2 3


# 1 "/usr/include/c++/10/x86_64-redhat-linux/bits/cpu_defines.h" 1 3
# 2617 "/usr/include/c++/10/x86_64-redhat-linux/bits/c++config.h" 2 3
# 50 "/usr/include/c++/10/cstddef" 2 3
# 1 "/usr/lib/gcc/x86_64-redhat-linux/10/include/stddef.h" 1 3 4
# 143 "/usr/lib/gcc/x86_64-redhat-linux/10/include/stddef.h" 3 4
typedef long int ptrdiff_t;
# 209 "/usr/lib/gcc/x86_64-redhat-linux/10/include/stddef.h" 3 4
typedef long unsigned int size_t;
# 415 "/usr/lib/gcc/x86_64-redhat-linux/10/include/stddef.h" 3 4
typedef struct {
  long long __max_align_ll __attribute__((__aligned__(__alignof__(long long))));
  long double __max_align_ld __attribute__((__aligned__(__alignof__(long double))));
# 426 "/usr/lib/gcc/x86_64-redhat-linux/10/include/stddef.h" 3 4
} max_align_t;






  typedef decltype(nullptr) nullptr_t;
# 51 "/usr/include/c++/10/cstddef" 2 3

extern "C++"
{

namespace std
{

  using ::max_align_t;
}
# 179 "/usr/include/c++/10/cstddef" 3
}
# 31 "RegistExt.h" 2




# 34 "RegistExt.h"
extern "C" {





void sqrtFunc(sqlite3_context *context, int argc, sqlite3_value **argv);
void logFunc(sqlite3_context *context, int argc, sqlite3_value **argv);
void expFunc(sqlite3_context *context, int argc, sqlite3_value **argv);
void powFunc(sqlite3_context *context, int argc, sqlite3_value **argv);



void CorrelStep(sqlite3_context *context, int argc, sqlite3_value **argv);
void CorrelFinal(sqlite3_context *context);
void SpCorrelStep(sqlite3_context *context, int argc, sqlite3_value **argv);
void SpCorrelFinal(sqlite3_context *context);



extern const sqlite3_api_routines *sqlite3_api;
extern sqlite3 *thisdb;




extern sqlite3_module histoModule;

int histoConnect(
  sqlite3 *db,
  void *pAux,
  int argc, const char *const*argv,
  sqlite3_vtab **ppVtab,
  char **pzErr
  );
int histoDisconnect(sqlite3_vtab *pVtab);
int histoOpen(sqlite3_vtab *p, sqlite3_vtab_cursor **ppCursor);
int histoClose(sqlite3_vtab_cursor *cur);
int histoNext(sqlite3_vtab_cursor *cur);
int histoColumn(sqlite3_vtab_cursor *cur, sqlite3_context *ctx, int i);
int histoRowid(sqlite3_vtab_cursor *cur, sqlite_int64 *pRowid);
int histoEof(sqlite3_vtab_cursor *cur);
int histoFilter(
  sqlite3_vtab_cursor *pVtabCursor,
  int idxNum, const char *idxStr,
  int argc, sqlite3_value **argv
  );
int histoBestIndex(sqlite3_vtab *tab, sqlite3_index_info *pIdxInfo);




extern sqlite3_module ratiohistoModule;

int ratiohistoConnect(
  sqlite3 *db,
  void *pAux,
  int argc, const char *const*argv,
  sqlite3_vtab **ppVtab,
  char **pzErr
  );
int ratiohistoDisconnect(sqlite3_vtab *pVtab);
int ratiohistoOpen(sqlite3_vtab *p, sqlite3_vtab_cursor **ppCursor);
int ratiohistoClose(sqlite3_vtab_cursor *cur);
int ratiohistoNext(sqlite3_vtab_cursor *cur);
int ratiohistoColumn(sqlite3_vtab_cursor *cur, sqlite3_context *ctx, int i);
int ratiohistoRowid(sqlite3_vtab_cursor *cur, sqlite_int64 *pRowid);
int ratiohistoEof(sqlite3_vtab_cursor *cur);
int ratiohistoFilter(
  sqlite3_vtab_cursor *pVtabCursor,
  int idxNum, const char *idxStr,
  int argc, sqlite3_value **argv
  );
int ratiohistoBestIndex(sqlite3_vtab *tab, sqlite3_index_info *pIdxInfo);



extern sqlite3_module meanhistoModule;

int meanhistoConnect(
  sqlite3 *db,
  void *pAux,
  int argc, const char *const*argv,
  sqlite3_vtab **ppVtab,
  char **pzErr
  );
int meanhistoDisconnect(sqlite3_vtab *pVtab);
int meanhistoOpen(sqlite3_vtab *p, sqlite3_vtab_cursor **ppCursor);
int meanhistoClose(sqlite3_vtab_cursor *cur);
int meanhistoNext(sqlite3_vtab_cursor *cur);
int meanhistoColumn(sqlite3_vtab_cursor *cur, sqlite3_context *ctx, int i);
int meanhistoRowid(sqlite3_vtab_cursor *cur, sqlite_int64 *pRowid);
int meanhistoEof(sqlite3_vtab_cursor *cur);
int meanhistoFilter(
  sqlite3_vtab_cursor *pVtabCursor,
  int idxNum, const char *idxStr,
  int argc, sqlite3_value **argv
  );
int meanhistoBestIndex(sqlite3_vtab *tab, sqlite3_index_info *pIdxInfo);






}



       
# 68 "RegistExt.cpp" 2



extern "C" {


sqlite3 *thisdb = 
# 74 "RegistExt.cpp" 3 4
                 __null
# 74 "RegistExt.cpp"
                     ;







int sqlite3_histograms_init(
  sqlite3 *db,
  char **pzErrMsg,
  const sqlite3_api_routines *pApi
  )
{
  int rc = 0;
  (void)pApi;;

  if (sqlite3_libversion_number()<3008012)
  {
    *pzErrMsg = sqlite3_mprintf("Histogram extension requires SQLite 3.8.12 or later");
    return 1;
  }
  rc = sqlite3_create_module(db, "HISTO", &histoModule, 0);
  rc = sqlite3_create_module(db, "RATIOHISTO", &ratiohistoModule, 0);
  rc = sqlite3_create_module(db, "MEANHISTO", &meanhistoModule, 0);


  sqlite3_create_function(db, "SQRT", 1, 5, 0, sqrtFunc, 0, 0);
  sqlite3_create_function(db, "LOG", 1, 5, 0, logFunc, 0, 0);
  sqlite3_create_function(db, "EXP", 1, 5, 0, expFunc, 0, 0);
  sqlite3_create_function(db, "POW", 2, 5, 0, powFunc, 0, 0);

  sqlite3_create_function(db, "CORREL", 2, 5, db, 
# 106 "RegistExt.cpp" 3 4
                                                          __null
# 106 "RegistExt.cpp"
                                                              , CorrelStep, CorrelFinal);
  sqlite3_create_function(db, "SPEARMANCORREL", 2, 5, db, 
# 107 "RegistExt.cpp" 3 4
                                                                  __null
# 107 "RegistExt.cpp"
                                                                      , SpCorrelStep, SpCorrelFinal);


  return rc;
}




}
