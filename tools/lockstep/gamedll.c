/* gamedll.dll: stands in for a game module (D2Game.dll) that calls the clock
 * functions through its own import table. gate_run mimics the 25 fps frame
 * gate: if (timeGetTime() - last >= 40) { frame++; last += 40; } Sleep(1); */
#include "common.h"

__declspec(dllexport) void gate_run(unsigned secs) {
    unsigned long long r0 = refnow(), end = r0 + (unsigned long long)secs * 10000000ULL;
    FILETIME c, e, k, u;
    DWORD t0 = timeGetTime(), last = t0, frame = 0, iters = 0;
    while (refnow() < end) {
        if (timeGetTime() - last >= 40) { frame++; last += 40; }
        Sleep(1);
        iters++;
    }
    DWORD t1 = timeGetTime();
    unsigned long long r1 = refnow();
    GetProcessTimes(GetCurrentProcess(), &c, &e, &k, &u);
    out("GATE frames=%u iters=%u ref_100ns=%u virt_ms=%u cpu_kernel_100ns=%u cpu_user_100ns=%u\r\n",
        frame, iters, (DWORD)(r1 - r0), t1 - t0, k.dwLowDateTime, u.dwLowDateTime);
}

BOOL WINAPI DllMain(HINSTANCE h, DWORD r, LPVOID x) { return TRUE; }
