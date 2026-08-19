#import <Cocoa/Cocoa.h>

// Switches the running process to an accessory application (no Dock icon,
// no Cmd+Tab entry). Dispatched onto the main queue so it is safe to call
// from any goroutine/thread and is correctly ordered relative to AppKit's
// own launch-time work on the main thread.
void awgproxy_hide_dock_icon(void) {
    dispatch_async(dispatch_get_main_queue(), ^{
        [[NSApplication sharedApplication] setActivationPolicy:NSApplicationActivationPolicyAccessory];
    });
}
