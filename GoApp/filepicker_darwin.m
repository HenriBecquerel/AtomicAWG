#import <Cocoa/Cocoa.h>
#include <string.h>
#include <stdlib.h>

// Presents the system NSOpenPanel and returns the chosen path as a
// heap-allocated C string (caller must free it via awgproxy_free_string),
// or NULL if the user cancelled.
static char *runOpenPanel(void) {
    NSOpenPanel *panel = [NSOpenPanel openPanel];
    panel.canChooseFiles = YES;
    panel.canChooseDirectories = NO;
    panel.allowsMultipleSelection = NO;
    panel.title = @"Выберите конфигурацию WireGuard";
    panel.message = @"Выберите файл .conf с настройками WireGuard или AmneziaWG";
    panel.prompt = @"Открыть";

    NSModalResponse response = [panel runModal];
    if (response == NSModalResponseOK && panel.URL != nil) {
        const char *path = panel.URL.fileSystemRepresentation;
        if (path != NULL) {
            return strdup(path);
        }
    }
    return NULL;
}

// AppKit UI must run on the main thread; runOpenPanel is dispatched there
// synchronously when called from any other thread.
char *awgproxy_choose_config_file(void) {
    if ([NSThread isMainThread]) {
        return runOpenPanel();
    }
    __block char *result = NULL;
    dispatch_sync(dispatch_get_main_queue(), ^{
        result = runOpenPanel();
    });
    return result;
}

void awgproxy_free_string(char *s) {
    free(s);
}
