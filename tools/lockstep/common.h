/* Shared helpers for the lockstep P0 probes. No CRT: output through WriteFile,
 * no 64-bit division (no compiler-rt on the build host). */
#include <windows.h>
#include <mmsystem.h>

static void out(const char *fmt, ...) {
    char buf[512];
    va_list ap;
    va_start(ap, fmt);
    int n = wvsprintfA(buf, fmt, ap);
    va_end(ap);
    DWORD w;
    WriteFile(GetStdHandle(STD_OUTPUT_HANDLE), buf, n, &w, NULL);
}

/* Host reference clock: Wine maps the precise system time to the host
 * realtime clock. Never patched by the hook. Units of 100 ns. */
static unsigned long long refnow(void) {
    FILETIME ft;
    GetSystemTimePreciseAsFileTime(&ft);
    return ((unsigned long long)ft.dwHighDateTime << 32) | ft.dwLowDateTime;
}
