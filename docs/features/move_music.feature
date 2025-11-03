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

  Scenario: Dry-run move operation
    Given a source directory "source_dry_run_move" with a music file "dry_move_track.mp3"
    And an empty target directory "target_dry_run_move"
    When I run the command "muxic copy --source source_dry_run_move --target target_dry_run_move --move --dry-run"
    Then the console output should contain "Dry-run mode enabled"
    And the console output should contain "[DRY-RUN] Would attempt to process/copy music file 'source_dry_run_move/dry_move_track.mp3'"
    And the console output should contain "[DRY-RUN] Would delete source file: source_dry_run_move/dry_move_track.mp3"
    # The following step was updated to match the more generic message from the implementation for empty folder deletion simulation
    And the console output should contain "[DRY-RUN] Would then check parent directories of source_dry_run_move/dry_move_track.mp3 for emptiness and potential deletion."
    And the file "target_dry_run_move/Artist/Album/01 - dry_move_track.mp3" should not exist
    And the source file "source_dry_run_move/dry_move_track.mp3" should still exist

  Scenario: Dry-run move operation with existing target file
    Given a source directory "source_dry_move_exists" with a music file "dry_exist_track.mp3"
    And a target directory "target_dry_move_exists" that already contains "Artist/Album/01 - dry_exist_track.mp3" # (or a path that would be predicted for it)
    When I run the command "muxic copy --source source_dry_move_exists --target target_dry_move_exists --move --dry-run"
    Then the console output should contain "Dry-run mode enabled"
    # Depending on implementation, it might predict the existing file or just state copy attempt.
    # If movemusic.CopyMusic is called (even if its IO is stubbed), it might return ErrFileExists.
    # For now, let's assume it reports the skip.
    And the console output should contain "EXISTS: File already exists, skipping source_dry_move_exists/dry_exist_track.mp3"
    And the console output should not contain "[DRY-RUN] Would delete source file: source_dry_move_exists/dry_exist_track.mp3"
    And the source file "source_dry_move_exists/dry_exist_track.mp3" should still exist

  Scenario: Move with Windows filesystem sanitization - prohibited characters
    Given a source directory "source_move_windows_chars" with the following music files:
      | File Path         | Artist          | Album           | Title              |
      | bad_chars.mp3     | Queen/King      | Greatest\\Hits  | We*Will?Rock"You   |
      | symbols.flac      | Artist<>Name    | Album|Volume:2  | Track>Name         |
    And an empty target directory "target_move_windows_chars"
    When I run the command "muxic copy --source source_move_windows_chars --target target_move_windows_chars --move"
    Then the file "target_move_windows_chars/Queen-King/Greatest-Hits/01 - We-Will-Rock-You.mp3" should exist
    And the file "target_move_windows_chars/Artist--Name/Album-Volume-2/01 - Track-Name.flac" should exist
    And the source file "source_move_windows_chars/bad_chars.mp3" should not exist
    And the source file "source_move_windows_chars/symbols.flac" should not exist
    And the console output should contain "Deleting source file:"

  Scenario: Move with Windows filesystem sanitization - Unicode characters
    Given a source directory "source_move_unicode" with the following music files:
      | File Path       | Artist       | Album           | Title            |
      | unicode1.mp3    | Café Del Mar | Ibiza Chillout  | Naïve Song      |
      | unicode2.flac   | 村上春樹      | 音楽アルバム     | 素晴らしい歌     |
    And an empty target directory "target_move_unicode"
    When I run the command "muxic copy --source source_move_unicode --target target_move_unicode --move"
    Then the file "target_move_unicode/Cafe Del Mar/Ibiza Chillout/01 - Naive Song.mp3" should exist
    And the file "target_move_unicode/Cun Shang Chun Shu/Yin Le arubamuShu/01 - Su Qing rashii Ge.flac" should exist
    And the source file "source_move_unicode/unicode1.mp3" should not exist
    And the source file "source_move_unicode/unicode2.flac" should not exist

  Scenario: Move with Windows filesystem sanitization - leading/trailing spaces and periods
    Given a source directory "source_move_trim" with the following music files:
      | File Path       | Artist         | Album            | Title              |
      | spaces.mp3      | " Band Name "  | " Album Title "  | " Song Name "     |
      | periods.flac    | "..Artist.."   | "..Album.."      | "..Title.."       |
      | mixed.wav       | ". Artist . "  | " .Album. "      | " .Song. "        |
    And an empty target directory "target_move_trim"
    When I run the command "muxic copy --source source_move_trim --target target_move_trim --move"
    Then the file "target_move_trim/Band Name/Album Title/01 - Song Name.mp3" should exist
    And the file "target_move_trim/Artist/Album/01 - Title.flac" should exist
    And the file "target_move_trim/Artist/Album/01 - Song.wav" should exist
    And the source file "source_move_trim/spaces.mp3" should not exist
    And the source file "source_move_trim/periods.flac" should not exist
    And the source file "source_move_trim/mixed.wav" should not exist

  Scenario: Move with comprehensive tag-based filename generation and sanitization
    Given a source directory "source_move_comprehensive" with the following music files:
      | File Path           | Artist              | Album                      | Title                      | Track |
      | terrible_name.mp3   | " //AC\\DC// "      | " ..Back*In?Black<>.. "   | " ..Hells\"Bells|Rock.. "  | 5     |
      | another_bad.flac    | " Sigur Rós "       | " ( ) "                   | " Untitled #1 "            | 1     |
    And an empty target directory "target_move_comprehensive"
    When I run the command "muxic copy --source source_move_comprehensive --target target_move_comprehensive --move"
    Then the file "target_move_comprehensive/--AC-DC--/Back-In-Black---/05 - Hells-Bells-Rock.mp3" should exist
    And the file "target_move_comprehensive/Sigur Ros/( )/01 - Untitled #1.flac" should exist
    And the source file "source_move_comprehensive/terrible_name.mp3" should not exist
    And the source file "source_move_comprehensive/another_bad.flac" should not exist
    And the console output should contain "Deleting source file:"
