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
