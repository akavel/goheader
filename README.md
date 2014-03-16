goheader
========

Tool for translating WinAPI C-type declarations/headers into their Go equivalent

suggested usage
---------------

The following commands:

	set MINGW_ROOT=c:\mingw-tdm32
	set PATH=%MINGW_ROOT%\bin;%PATH%
	mingw32-gcc -E -D UNICODE -D _UNICODE %MINGW_ROOT%\include\windows.h > winapi_unicode.h
	
	goheader -p winapi _RPC_MESSAGE _RPC_SYNTAX_IDENTIFIER _GUID _RPC_VERSION < winapi_unicode.h | gofmt

should give a result like below:

	package winapi
	
	type _GUID struct { // typedef struct _GUID {
		Data1 uint32  // unsigned long Data1;
		Data2 uint16  // unsigned short Data2;
		Data3 uint16  // unsigned short Data3;
		Data4 [8]byte // unsigned char Data4[8];
	}
	type _RPC_VERSION struct { // typedef struct _RPC_VERSION {
		MajorVersion uint16 // unsigned short MajorVersion;
		MinorVersion uint16 // unsigned short MinorVersion;
	}
	type _RPC_SYNTAX_IDENTIFIER struct { // typedef struct _RPC_SYNTAX_IDENTIFIER {
		SyntaxGUID    _GUID        // GUID SyntaxGUID;
		SyntaxVersion _RPC_VERSION // RPC_VERSION SyntaxVersion;
	}
	type _RPC_MESSAGE struct { // typedef struct _RPC_MESSAGE {
		Handle                  uintptr                 // HANDLE Handle;
		DataRepresentation      uint32                  // unsigned long DataRepresentation;
		Buffer                  uintptr                 // void *Buffer;
		BufferLength            uint32                  // unsigned int BufferLength;
		ProcNum                 uint32                  // unsigned int ProcNum;
		TransferSyntax          *_RPC_SYNTAX_IDENTIFIER // PRPC_SYNTAX_IDENTIFIER TransferSyntax;
		RpcInterfaceInformation uintptr                 // void *RpcInterfaceInformation;
		ReservedForRuntime      uintptr                 // void *ReservedForRuntime;
		ManagerEpv              uintptr                 // void *ManagerEpv;
		ImportContext           uintptr                 // void *ImportContext;
		RpcFlags                uint32                  // unsigned long RpcFlags;
	}

caveats
-------

  - doesn't parse function pointers nor function declarations as of now.
  - doesn't parse unions, enums, nested structs.
  - it makes assumptions on width of C `int` etc, you can tweak them in `translatePrimitive()` for your needs.
  - may eat your homework, etc.

In cases above (except the homework), you may still call it without any arguments, and pipe the `.h` through it; it will dump everything what it could, and there may still be some partial work done for you already in the output.


license
-------

Copyright (c) 2014 by Mateusz Czapliński <czapkofan@gmail.com>

The code of goheader is released by me to Public Domain;

If that does not work for you, I hereby allow copying this crap (i.e. the full sourcecode of goheader) under whichever license you prefer from the list below:

  * [WTFPL](https://en.wikipedia.org/wiki/WTFPL),
  * BSD 2-clause or 3-clause,
  * "MIT/X" or "MIT/Expat" or I don't know whatever else.
