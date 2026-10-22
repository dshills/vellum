"""Add failure-only context to a copied, previously compiled diagnostic consumer."""
def replace_once(text, old, new):
    if text.count(old) != 1 or new in text:
        raise ValueError('failure-context source drift')
    return text.replace(old, new)


def patch_main(text):
    return replace_once(text, '\t\t\tif !good(state) {\n\t\t\t\treturn errors.New("foreground/configuration changed during cohort")\n\t\t\t}', '\t\t\tif err := cohortStateError(out, i, state, good(state), func() map[string]any { return observe(3) }); err != nil {\n\t\t\t\treturn err\n\t\t\t}')


def patch_observer(text):
    return replace_once(text, '  NSData *data=[NSJSONSerialization dataWithJSONObject:d', '''  if(phase==3){
   // Only after an input has already failed; no added work on admitted inputs.
   NSRunningApplication *front=NSWorkspace.sharedWorkspace.frontmostApplication;
   NSMutableDictionary *detail=[NSMutableDictionary dictionaryWithDictionary:d];
   detail[@"front_bundle"]=front.bundleIdentifier?:@"";
   detail[@"front_name"]=front.localizedName?:@"";
   detail[@"owned_window_count"]=@(NSApp.windows.count);
   d=detail;
  }
  NSData *data=[NSJSONSerialization dataWithJSONObject:d''')
