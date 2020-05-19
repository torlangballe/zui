#ifndef BRIDGE_H
#define BRIDGE_H

struct WinScale {
    int WinID;
    int Scale;
};

long long GetNetworkTrafficBytes();
char     *GetAppContentPath();
char     *GetDeviceUniqueID();
int       GetFirstWindowIDForPID(int pid); // returns 0 if none
char     *GetWindowIDForAppAndTitle(char *app, char *title); 
int       DoesWindowWithTitleExists(char *title); // 0 or 1
char     *GetIPAddressForEthernet();
char     *GetDeviceIdentifier();
int       GetMainScreenScale();

#endif
