Feature: Move Music Files
  As a user of the muxic CLI,
  I want to be able to move music files from a source directory to a target directory,
  so that I can organize my music library and remove the original files.

  Scenario: Basic move operation
    Given a source directory "source_to_move" with a music file "track_to_move.mp3"
    And an empty target directory "target_for_move"
    When I run the command "muxic copy --source source_to_move --target target_for_move --move"
    Then the file "target_for_move/Artist/Album/01 - track_to_move.mp3" should exist
    And the source file "source_to_move/track_to_move.mp3" should not exist
    And the console output should contain "Deleting source file: source_to_move/track_to_move.mp3"

  Scenario: Move operation with verbose logging
    Given a source directory "source_to_move_verbose" with a music file "verbose_move.mp3"
    And an empty target directory "target_for_move_verbose"
    When I run the command "muxic copy --source source_to_move_verbose --target target_for_move_verbose --move --verbose"
    Then the file "target_for_move_verbose/Artist/Album/01 - verbose_move.mp3" should exist
    And the source file "source_to_move_verbose/verbose_move.mp3" should not exist
    And the console output should contain "Copying file: source_to_move_verbose/verbose_move.mp3"
    And the console output should contain "Finished: target_for_move_verbose/Artist/Album/01 - verbose_move.mp3"
    And the console output should contain "Deleting source file: source_to_move_verbose/verbose_move.mp3"

  Scenario: Move operation and empty source subfolder cleanup
    Given a source directory "source_empty_folder" with a music file "sub/track_in_sub.mp3"
    And an empty target directory "target_empty_folder"
    When I run the command "muxic copy --source source_empty_folder --target target_empty_folder --move"
    Then the file "target_empty_folder/Artist/Album/01 - track_in_sub.mp3" should exist
    And the source file "source_empty_folder/sub/track_in_sub.mp3" should not exist
    And the source directory "source_empty_folder/sub" should not exist
    And the console output should contain "Deleting empty source folder: source_empty_folder/sub"

  Scenario: Move operation when target file already exists
    Given a source directory "source_move_exists" with a music file "exist_track.mp3"
    And a target directory "target_move_exists" that already contains "Artist/Album/01 - exist_track.mp3"
    When I run the command "muxic copy --source source_move_exists --target target_move_exists --move"
    Then the console output should contain "EXISTS: File already exists, skipping source_move_exists/exist_track.mp3"
    And the source file "source_move_exists/exist_track.mp3" should still exist
    And the file "target_move_exists/Artist/Album/01 - exist_track.mp3" should not have been modified recently

  Scenario: Attempt to move a file that is identical in source and target (already organized)
    Given a source directory "source_identical" containing "Artist/Album/01 - same_song.mp3"
    And the target directory is "source_identical" (source and target are the same)
    When I run the command "muxic copy --source source_identical --target source_identical --move"
    Then the file "source_identical/Artist/Album/01 - same_song.mp3" should exist
    And the console output should not contain "Deleting source file: source_identical/Artist/Album/01 - same_song.mp3"
    And the console output should likely contain "EXISTS: File already exists" or similar, indicating it was skipped.
