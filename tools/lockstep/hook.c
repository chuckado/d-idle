/* fakeijl.dll: stand-in proxy DLL for the P0 probe. The test exe imports
 * it (like Game.exe imports IJL11.DLL), so Windows loads it and runs
 * DllMain before the exe's entry point. DllMain patches the import
 * address table slots of the listed modules for the clock functions with
 * versions that return virtual = base + (real - base) * k.
 *
 * Env: LOCKSTEP_KPCT   k in percent (100 = null hook, 200 = 2x); unset = no patch
 *      LOCKSTEP_SLEEP  1 = Sleep(ms) becomes Sleep(ms / k), default 0 (raw)
 *      LOCKSTEP_MODS   comma list of module names to patch; default
 *                      "exe,gamedll.dll" ("exe" is the main module)
 */
#include "common.h"

static unsigned kfix;     /* k in 16.16 fixed point */
static int scale_sleep;
static DWORD (WINAPI *r_tgt)(void);
static DWORD (WINAPI *r_gtc)(void);
static ULONGLONG (WINAPI *r_gtc64)(void);
static BOOL (WINAPI *r_qpc)(LARGE_INTEGER *);
static void (WINAPI *r_sleep)(DWORD);
static DWORD b_tgt, b_gtc;
static ULONGLONG b_gtc64;
static LONGLONG b_qpc;

static DWORD WINAPI h_tgt(void) { DWORD r = r_tgt(); return b_tgt + (DWORD)(((unsigned long long)(DWORD)(r - b_tgt) * kfix) >> 16); }
static DWORD WINAPI h_gtc(void) { DWORD r = r_gtc(); return b_gtc + (DWORD)(((unsigned long long)(DWORD)(r - b_gtc) * kfix) >> 16); }
static ULONGLONG WINAPI h_gtc64(void) { ULONGLONG r = r_gtc64(); return b_gtc64 + (((r - b_gtc64) * kfix) >> 16); }
static BOOL WINAPI h_qpc(LARGE_INTEGER *p) {
    LARGE_INTEGER q;
    BOOL ok = r_qpc(&q);
    p->QuadPart = b_qpc + (LONGLONG)(((unsigned long long)(q.QuadPart - b_qpc) * kfix) >> 16);
    return ok;
}
static void WINAPI h_sleep(DWORD ms) {
    if (scale_sleep && ms < 65536) ms = (ms << 16) / kfix;
    r_sleep(ms);
}

static int patch(HMODULE m, const char *name, void *fn) {
    BYTE *base = (BYTE *)m;
    IMAGE_DOS_HEADER *dh = (IMAGE_DOS_HEADER *)base;
    IMAGE_NT_HEADERS *nt = (IMAGE_NT_HEADERS *)(base + dh->e_lfanew);
    IMAGE_DATA_DIRECTORY dd = nt->OptionalHeader.DataDirectory[IMAGE_DIRECTORY_ENTRY_IMPORT];
    if (!dd.VirtualAddress) return 0;
    int n = 0;
    for (IMAGE_IMPORT_DESCRIPTOR *d = (IMAGE_IMPORT_DESCRIPTOR *)(base + dd.VirtualAddress); d->Name; d++) {
        DWORD *oft = (DWORD *)(base + (d->OriginalFirstThunk ? d->OriginalFirstThunk : d->FirstThunk));
        void **iat = (void **)(base + d->FirstThunk);
        for (; *oft; oft++, iat++) {
            if (*oft & 0x80000000) continue; /* ordinal import */
            const char *nm = (const char *)(base + *oft + 2);
            if (lstrcmpA(nm, name)) continue;
            DWORD old;
            VirtualProtect(iat, 4, PAGE_READWRITE, &old);
            *iat = fn;
            VirtualProtect(iat, 4, old, &old);
            n++;
        }
    }
    return n;
}

static int patch_all(HMODULE m, const char *label) {
    int n = 0;
    n += patch(m, "timeGetTime", h_tgt);
    n += patch(m, "GetTickCount", h_gtc);
    n += patch(m, "GetTickCount64", h_gtc64);
    n += patch(m, "QueryPerformanceCounter", h_qpc);
    n += patch(m, "Sleep", h_sleep);
    out("HOOK module=%s patched_slots=%d\r\n", label, n);
    return n;
}

__declspec(dllexport) void ijl_stub(void) {}

BOOL WINAPI DllMain(HINSTANCE h, DWORD reason, LPVOID res) {
    if (reason != DLL_PROCESS_ATTACH) return TRUE;
    char kbuf[32];
    if (!GetEnvironmentVariableA("LOCKSTEP_KPCT", kbuf, sizeof kbuf)) {
        out("HOOK none (LOCKSTEP_KPCT unset)\r\n");
        return TRUE;
    }
    unsigned pct = 0;
    for (char *p = kbuf; *p >= '0' && *p <= '9'; p++) pct = pct * 10 + (*p - '0');
    kfix = (pct << 16) / 100;
    char sb[8];
    scale_sleep = GetEnvironmentVariableA("LOCKSTEP_SLEEP", sb, sizeof sb) && sb[0] == '1';
    HMODULE k32 = GetModuleHandleA("kernel32.dll"), wm = LoadLibraryA("winmm.dll");
    r_tgt = (void *)GetProcAddress(wm, "timeGetTime");
    r_gtc = (void *)GetProcAddress(k32, "GetTickCount");
    r_gtc64 = (void *)GetProcAddress(k32, "GetTickCount64");
    r_qpc = (void *)GetProcAddress(k32, "QueryPerformanceCounter");
    r_sleep = (void *)GetProcAddress(k32, "Sleep");
    LARGE_INTEGER q;
    b_tgt = r_tgt(); b_gtc = r_gtc(); b_gtc64 = r_gtc64(); r_qpc(&q); b_qpc = q.QuadPart;
    out("HOOK k_pct=%u kfix=0x%x sleep_scaled=%d\r\n", pct, kfix, scale_sleep);
    char mods[128];
    if (!GetEnvironmentVariableA("LOCKSTEP_MODS", mods, sizeof mods)) lstrcpyA(mods, "exe,gamedll.dll");
    for (char *p = mods; *p;) {
        char *e = p;
        while (*e && *e != ',') e++;
        char save = *e;
        *e = 0;
        HMODULE m = lstrcmpA(p, "exe") ? GetModuleHandleA(p) : GetModuleHandleA(NULL);
        if (m) patch_all(m, p); else out("HOOK module=%s not loaded\r\n", p);
        *e = save;
        p = save ? e + 1 : e;
    }
    return TRUE;
}
