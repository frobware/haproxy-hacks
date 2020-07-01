/*
RegistExt.cpp, Robert Oeffner 2018

The MIT License (MIT)

Copyright (c) 2017 Robert Oeffner

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.




Compile on Windows linking with whole program optimisation:

cl /Ox /EHsc /GL /Fohelpers.obj /c helpers.cpp  ^
 && cl /Ox /EHsc /GL /FoSQLiteExt.obj /c SQLiteExt.cpp ^
 && cl /Ox /EHsc /GL /Foratiohistogram.obj /c ratiohistogram.cpp ^
 && cl /Ox /EHsc /GL /Fohistogram.obj /c histogram.cpp ^
 && cl /Ox /EHsc /GL /Fomeanhistogram.obj /c meanhistogram.cpp ^
 && cl /Ox /EHsc /GL /FoRegistExt.obj /c RegistExt.cpp ^
 && link /DLL /LTCG /OUT:histograms.dll helpers.obj SQLiteExt.obj RegistExt.obj meanhistogram.obj histogram.obj ratiohistogram.obj

With debug info:

cl /Fohelpers.obj /c helpers.cpp /DDEBUG /ZI /EHsc ^
 && cl /FoSQLiteExt.obj /c SQLiteExt.cpp /DDEBUG  /ZI /EHsc ^
 && cl /Foratiohistogram.obj /c ratiohistogram.cpp /DDEBUG  /ZI /EHsc ^
 && cl /Fomeanhistogram.obj /c meanhistogram.cpp /DDEBUG  /ZI /EHsc ^
 && cl /Fohistogram.obj /c histogram.cpp /DDEBUG  /ZI /EHsc ^
 && cl /FoRegistExt.obj /c RegistExt.cpp  /DDEBUG  /ZI /EHsc ^
 && link /DLL /DEBUG /debugtype:cv /OUT:histograms.dll helpers.obj SQLiteExt.obj meanhistogram.obj RegistExt.obj histogram.obj ratiohistogram.obj

 
Compile on Linux:

 g++ -fPIC -lm -shared histogram.cpp helpers.cpp meanhistogram.cpp ratiohistogram.cpp RegistExt.cpp -o libhistograms.so

 From the sqlite commandline load the extension

 on Windows
 sqlite> .load histograms.dll
 sqlite>
 
 on Linux
 sqlite> .load ./histograms.so
 sqlite>

*/


#include "RegistExt.h"


#ifdef __cplusplus
extern "C" {
#endif


SQLITE_EXTENSION_INIT1



sqlite3 *thisdb = NULL;

#ifdef _WIN32
__declspec(dllexport)
#endif
/* The built library file name excluding its file extension must be part of the 
 function name below as documented on http://www.sqlite.org/loadext.html
*/
int sqlite3_histograms_init( // always use lower case
  sqlite3 *db,
  char **pzErrMsg,
  const sqlite3_api_routines *pApi
  )
{
  int rc = SQLITE_OK;
  SQLITE_EXTENSION_INIT2(pApi);
#ifndef SQLITE_OMIT_VIRTUALTABLE
  if (sqlite3_libversion_number()<3008012)
  {
    *pzErrMsg = sqlite3_mprintf("Histogram extension requires SQLite 3.8.12 or later");
    return SQLITE_ERROR;
  }
  rc = sqlite3_create_module(db, "HISTO", &histoModule, 0);
  rc = sqlite3_create_module(db, "RATIOHISTO", &ratiohistoModule, 0);
  rc = sqlite3_create_module(db, "MEANHISTO", &meanhistoModule, 0);

  // 3. parameter is the number of arguments the functions take 
  sqlite3_create_function(db, "SQRT", 1, SQLITE_ANY, 0, sqrtFunc, 0, 0);
  sqlite3_create_function(db, "LOG", 1, SQLITE_ANY, 0, logFunc, 0, 0);
  sqlite3_create_function(db, "EXP", 1, SQLITE_ANY, 0, expFunc, 0, 0);
  sqlite3_create_function(db, "POW", 2, SQLITE_ANY, 0, powFunc, 0, 0);

  sqlite3_create_function(db, "CORREL", 2, SQLITE_ANY, db, NULL, CorrelStep, CorrelFinal);
  sqlite3_create_function(db, "SPEARMANCORREL", 2, SQLITE_ANY, db, NULL, SpCorrelStep, SpCorrelFinal);

#endif
  return rc;
}



#ifdef __cplusplus
}
#endif
