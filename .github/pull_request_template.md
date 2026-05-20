## Issue

Which issue does this PR address?

Fixes ghazlabs/wa-scheduler#19

## Root Cause

The API returned a generic error response when a message ID was not found, making it difficult for clients to identify the actual issue clearly.

## Changes

- Added a more declarative and explicit error message for missing message IDs.
- Improved error handling for `core.ErrMessageNotFound`.
- Returned a proper `404 Not Found` response with a clearer error description.


### Manual Testing

- Request message detail using a non-existent message ID.
- Verify the API returns:
  - HTTP `404 Not Found`
  - Clear error message indicating the message ID was not found.
