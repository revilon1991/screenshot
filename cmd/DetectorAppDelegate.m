#import <Foundation/Foundation.h>
#import <AppKit/AppKit.h>
#import <pwd.h>
#import <DetectorAppDelegate.h>

void queryWire(void *pathPointer, void *createdAtPointer);
void ui();
void openPreferences();
void openHelp();
void savePreferences();

@implementation DetectorAppDelegate
@synthesize queryResults;

- (void)applicationDidFinishLaunching:(NSNotification *)aNotification
{
    ui();

    query = [[NSMetadataQuery alloc] init];
    NSNotificationCenter* defaultCenter = [NSNotificationCenter defaultCenter];

    [defaultCenter addObserver:self
                      selector:@selector(queryUpdated:)
                          name:NSMetadataQueryDidStartGatheringNotification
                        object:query];
    [defaultCenter addObserver:self
                      selector:@selector(queryUpdated:)
                          name:NSMetadataQueryDidUpdateNotification
                        object:query];
    [defaultCenter addObserver:self
                      selector:@selector(queryUpdated:)
                          name:NSMetadataQueryDidFinishGatheringNotification
                        object:query];

    [query setDelegate:self];
    [query setPredicate:[NSPredicate predicateWithFormat:@"kMDItemIsScreenCapture = 1"]];

    NSMutableArray *pathURLs = [self accessToFolder];
    [query setSearchScopes:pathURLs];

    [query startQuery];
}

- (void)applicationWillTerminate:(NSNotification *)notification {
    [query stopQuery];
    [query setDelegate:nil];
    (void)([query release]), query = nil;

    [self setQueryResults:nil];
}

- (void)queryUpdated:(NSNotification *)note {
    [self setQueryResults:[query results]];

    for(NSMetadataItem *item in [query results])
    {
        NSString *createdAt = [item valueForAttribute:NSMetadataItemFSCreationDateKey];
        NSString *path = [item valueForAttribute:NSMetadataItemPathKey];

        queryWire(path, createdAt);
    }
}

- (NSMutableArray *)accessToFolder {
    const char *home = getpwuid(getuid())->pw_dir;
    NSString *path = [[NSFileManager defaultManager] stringWithFileSystemRepresentation:home
                                                                                 length:strlen(home)];
    NSString *realHomeDirectory = [[NSURL fileURLWithPath:path isDirectory:YES] path];

    NSMutableArray *pathURLs = [NSMutableArray array];

    [pathURLs addObject:[NSURL fileURLWithPath:[NSString stringWithFormat:@"%@/Desktop", realHomeDirectory] isDirectory:true]];
    [pathURLs addObject:[NSURL fileURLWithPath:[NSString stringWithFormat:@"%@/Pictures", realHomeDirectory] isDirectory:true]];

    return pathURLs;
}

- (void)openPreferencesSel {
    openPreferences();
}

- (void)openHelpSel {
    openHelp();
}

- (void)savePreferencesSel {
    savePreferences();
}
@end
