Feature: Copy Music Files
  As a user of the muxic CLI,
  I want to be able to copy music files from a source directory to a target directory,
  so that I can organize my music library.

  Scenario: Basic copy operation
    Given a source directory "source_music" with the following music files:
      | File Path        | Artist  | Album   | Title   |
      | track1.mp3       | ArtistA | AlbumX  | Song 1  |
      | subfolder/track2.wav | ArtistB | AlbumY  | Song 2  |
    And an empty target directory "target_music"
    When I run the command "muxic copy --source source_music --target target_music"
    Then the file "target_music/ArtistA/AlbumX/01 - Song 1.mp3" should exist
    And the file "target_music/ArtistB/AlbumY/01 - Song 2.wav" should exist
    And the source file "source_music/track1.mp3" should still exist
    And the source file "source_music/subfolder/track2.wav" should still exist

  Scenario: Copy with verbose logging
    Given a source directory "source_music" with a music file "track1.mp3"
    And an empty target directory "target_music"
    When I run the command "muxic copy --source source_music --target target_music --verbose"
    Then the file "target_music/Artist/Album/01 - TrackTitle.mp3" should exist
    And the console output should contain "Copying file: source_music/track1.mp3"
    And the console output should contain "Finished: target_music/Artist/Album/01 - TrackTitle.mp3"

  Scenario: Target directory does not exist
    Given a source directory "source_music" with a music file "track1.mp3"
    And the target directory "new_target_music" does not exist
    When I run the command "muxic copy --source source_music --target new_target_music"
    Then the directory "new_target_music" should be created
    And the file "new_target_music/Artist/Album/01 - TrackTitle.mp3" should exist

  Scenario: File already exists in target directory
    Given a source directory "source_music" with a music file "track1.mp3"
    And a target directory "target_music" that already contains "Artist/Album/01 - TrackTitle.mp3"
    When I run the command "muxic copy --source source_music --target target_music"
    Then the console output should contain "EXISTS: File already exists, skipping source_music/track1.mp3"
    And the file "target_music/Artist/Album/01 - TrackTitle.mp3" should not have been modified recently

  Scenario: Copying multiple file types
    Given a source directory "source_music" with the following music files:
      | File Path        |
      | song.mp3         |
      | audio.flac       |
      | sound.m4a        |
      | music.wav        |
    And an empty target directory "target_music"
    When I run the command "muxic copy --source source_music --target target_music"
    Then the file "target_music/Artist/Album/01 - song.mp3" should exist
    Then the file "target_music/Artist/Album/01 - audio.flac" should exist
    Then the file "target_music/Artist/Album/01 - sound.m4a" should exist
    Then the file "target_music/Artist/Album/01 - music.wav" should exist

  Scenario: Source directory does not exist
    Given a source directory "non_existent_source" that does not exist
    And an empty target directory "target_music"
    When I run the command "muxic copy --source non_existent_source --target target_music"
    Then the command should fail
    And the console output should contain "Error accessing path" or "Error walking the path"

  Scenario: Dry-run copy operation
    Given a source directory "source_dry_run_copy" with a music file "dry_copy_track.mp3"
    And an empty target directory "target_dry_run_copy"
    When I run the command "muxic copy --source source_dry_run_copy --target target_dry_run_copy --dry-run"
    Then the console output should contain "[DRY-RUN] Would attempt to process/copy music file 'source_dry_run_copy/dry_copy_track.mp3'"
    And the console output should contain "Dry-run mode enabled"
    And the file "target_dry_run_copy/Artist/Album/01 - dry_copy_track.mp3" should not exist
    And the source file "source_dry_run_copy/dry_copy_track.mp3" should still exist
    And the target directory "target_dry_run_copy" should remain empty or not be created if it didn't initially exist (beyond its base)

  Scenario: Dry-run copy creating a new target directory
    Given a source directory "source_dry_run_new_target" with a music file "new_target_dry.mp3"
    And the target directory "target_dry_run_mkdir" does not exist
    When I run the command "muxic copy --source source_dry_run_new_target --target target_dry_run_mkdir --dry-run"
    Then the console output should contain "[DRY-RUN] Would create target folder: target_dry_run_mkdir"
    And the console output should contain "[DRY-RUN] Would attempt to process/copy music file 'source_dry_run_new_target/new_target_dry.mp3'"
    And the directory "target_dry_run_mkdir" should not be created
