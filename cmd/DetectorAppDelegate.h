#import <Foundation/Foundation.h>
#import <AppKit/AppKit.h>
#import <pwd.h>

@interface DetectorAppDelegate : NSObject <NSApplicationDelegate, NSMetadataQueryDelegate> {
@private
    NSMetadataQuery *query;
}

@property (nonatomic, copy) NSArray *queryResults;
@end
