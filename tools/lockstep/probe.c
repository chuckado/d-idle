/* probe.exe: P0 test program. PROBE_MODE = clocks | sleep | gate, PROBE_SECS.
 * Imports fakeijl.dll (hook stand-in) and gamedll.dll (module with its own
 * clock imports). Prints raw numbers; run.sh does the arithmetic. */
#include "common.h"

__declspec(dllimport) void ijl_stub(void);
__declspec(dllimport) void gate_run(unsigned secs);

static unsigned envnum(const char *n, unsigned d) {
    char b[16];
    if (!GetEnvironmentVariableA(n, b, sizeof b)) return d;
    unsigned v = 0;
    for (char *p = b; *p >= '0' && *p <= '9'; p++) v = v * 10 + (*p - '0');
    return v;
}
static int modeis(const char *m) {
    char b[16];
    return GetEnvironmentVariableA("PROBE_MODE", b, sizeof b) && !lstrcmpA(b, m);
}

static void clocks(unsigned secs) {
    LARGE_INTEGER f, q0, q;
    QueryPerformanceFrequency(&f);
    out("QPF %u\r\n", f.LowPart);
    /* resolution: spin 1 s of reference time, count value changes and the
     * smallest nonzero step (ticks of that API's own unit) */
    {
        unsigned long long end = refnow() + 10000000ULL;
        DWORD pt = timeGetTime(), pg = GetTickCount(), mt = ~0u, mg = ~0u, ct = 0, cg = 0;
        LARGE_INTEGER pq; QueryPerformanceCounter(&pq);
        unsigned long long mq = ~0ULL, cq = 0;
        while (refnow() < end) {
            DWORD t = timeGetTime(), g = GetTickCount();
            QueryPerformanceCounter(&q);
            if (t != pt) { if (t - pt < mt) mt = t - pt; ct++; pt = t; }
            if (g != pg) { if (g - pg < mg) mg = g - pg; cg++; pg = g; }
            if (q.QuadPart != pq.QuadPart) { unsigned long long d = q.QuadPart - pq.QuadPart; if (d < mq) mq = d; cq++; pq = q; }
        }
        out("RES tgt_changes=%u tgt_minstep=%u gtc_changes=%u gtc_minstep=%u qpc_changes=%u qpc_minstep=%u\r\n",
            ct, mt, cg, mg, (DWORD)cq, (DWORD)mq);
    }
    unsigned long long r0 = refnow();
    DWORD t0 = timeGetTime(), g0 = GetTickCount();
    ULONGLONG G0 = GetTickCount64();
    QueryPerformanceCounter(&q0);
    unsigned n = secs * 20; /* a sample every 50 ms */
    for (unsigned i = 0; i < n; i++) {
        Sleep(50);
        QueryPerformanceCounter(&q);
        out("S %u %u %u %u %u\r\n", (DWORD)(refnow() - r0), timeGetTime() - t0, GetTickCount() - g0,
            (DWORD)(GetTickCount64() - G0), (DWORD)(q.QuadPart - q0.QuadPart));
    }
}

static void sleeps(void) {
    unsigned long long a;
    for (int phase = 0; phase < 2; phase++) {
        if (phase) timeBeginPeriod(1);
        for (int j = 0; j < 2; j++) {
            DWORD ms = j ? 100 : 1, reps = j ? 10 : 200;
            unsigned long long tot = 0, mn = ~0ULL, mx = 0;
            for (DWORD i = 0; i < reps; i++) {
                a = refnow(); Sleep(ms); a = refnow() - a;
                tot += a; if (a < mn) mn = a; if (a > mx) mx = a;
            }
            out("SLEEP period1=%d ms=%u reps=%u total_100ns=%u min_100ns=%u max_100ns=%u\r\n",
                phase, ms, reps, (DWORD)tot, (DWORD)mn, (DWORD)mx);
        }
    }
    timeEndPeriod(1);
}

int mainCRTStartup(void) {
    ijl_stub();
    unsigned secs = envnum("PROBE_SECS", 10);
    if (modeis("clocks")) clocks(secs);
    else if (modeis("sleep")) sleeps();
    else if (modeis("gate")) gate_run(secs);
    ExitProcess(0);
    return 0;
}
