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


  Scenario: Windows filesystem sanitization - prohibited characters
    Given a source directory "source_windows_chars" with the following music files:
      | File Path             | Artist        | Album         | Title            |
      | track1.mp3            | AC/DC         | Back\\Black   | Hells*Bells     |
      | track2.flac           | Artist<Name>  | Album?Mix     | Song"Title"     |
      | track3.wav            | Band\|Name    | Album:Vol1    | Track>Name      |
    And an empty target directory "target_windows_chars"
    When I run the command "muxic copy --source source_windows_chars --target target_windows_chars"
    Then the file "target_windows_chars/AC-DC/Back-Black/01 - Hells-Bells.mp3" should exist
    And the file "target_windows_chars/Artist-Name-/Album-Mix/01 - Song-Title-.flac" should exist
    And the file "target_windows_chars/Band-Name/Album-Vol1/01 - Track-Name.wav" should exist
    And the console output should contain "Finished:"

  Scenario: Windows filesystem sanitization - Unicode and non-ASCII characters
    Given a source directory "source_unicode" with the following music files:
      | File Path         | Artist      | Album         | Title           |
      | track1.mp3        | Björk       | Médulla       | Öll Birtan     |
      | track2.flac       | Sigur Rós   | Ágætis byrjun | Svefn-g-englar |
      | track3.wav        | 中文歌手     | 专辑名称       | 歌曲标题        |
    And an empty target directory "target_unicode"
    When I run the command "muxic copy --source source_unicode --target target_unicode"
    Then the file "target_unicode/Bjork/Medulla/01 - Oll Birtan.mp3" should exist
    And the file "target_unicode/Sigur Ros/Agaetis Byrjun/01 - Svefn-G-Englar.flac" should exist
    And the file "target_unicode/Zhong Wen Ge Shou/Zhuan Ji Ming Cheng/01 - Ge Qu Biao Ti.wav" should exist

  Scenario: Windows filesystem sanitization - leading and trailing periods and spaces
    Given a source directory "source_trim_chars" with the following music files:
      | File Path         | Artist        | Album           | Title             |
      | track1.mp3        | " Artist "    | " Album Name "  | " Song Title "   |
      | track2.flac       | "..Artist.."  | "..Album.Name.."| "..Song.Title.." |
      | track3.wav        | ". Artist ."  | ". Album ."     | ". Title ."      |
    And an empty target directory "target_trim_chars"
    When I run the command "muxic copy --source source_trim_chars --target target_trim_chars"
    Then the file "target_trim_chars/Artist/Album Name/01 - Song Title.mp3" should exist
    And the file "target_trim_chars/Artist/Album.Name/01 - Song.Title.flac" should exist  
    And the file "target_trim_chars/Artist/Album/01 - Title.wav" should exist

  Scenario: Filename generation from tag metadata with sanitization
    Given a source directory "source_tag_metadata" with the following music files:
      | File Path         | Artist          | Album               | Title                  | Track |
      | badly_named.mp3   | The/Beatles     | Sgt. Pepper's*Club  | Lucy<In>The"Sky:Diamonds | 8     |
      | random.flac       | Daft Punk       | Random Access Memories | Get Lucky ft. Pharrell   | 3     |
      | file.wav          | ". Radiohead ." | "..OK Computer.."   | " Paranoid Android "      | 1     |
    And an empty target directory "target_tag_metadata"
    When I run the command "muxic copy --source source_tag_metadata --target target_tag_metadata"
    Then the file "target_tag_metadata/The-Beatles/Sgt. Pepper's-Club/08 - Lucy-In-The-Sky-Diamonds.mp3" should exist
    And the file "target_tag_metadata/Daft Punk/Random Access Memories/03 - Get Lucky Ft. Pharrell.flac" should exist
    And the file "target_tag_metadata/Radiohead/OK Computer/01 - Paranoid Android.wav" should exist
    And the console output should contain "Finished:"

  Scenario: Complex sanitization with multiple issues combined  
    Given a source directory "source_complex_sanitization" with the following music files:
      | File Path       | Artist               | Album                    | Title                       |
      | messy.mp3       | " //Mötley\\Crüe// " | " ..Shout*At?The<Devil>.. " | " ..Girls,\"Girls\",Girls.. " |
    And an empty target directory "target_complex_sanitization"
    When I run the command "muxic copy --source source_complex_sanitization --target target_complex_sanitization"
    Then the file "target_complex_sanitization/--Motley-Crue--/Shout-At-The-Devil-/01 - Girls,-Girls-,Girls.mp3" should exist

  Scenario: Complete tag-based filename generation with Windows filesystem compatibility
    Given a source directory "source_complete_tags" with the following music files with detailed metadata:
      | File Path          | Artist                  | Album                        | Title                         | Track | Genre      | Year |
      | originalname1.mp3  | The Beatles             | Sgt. Pepper's Lonely Hearts Club Band | Lucy In The Sky With Diamonds | 8     | Rock       | 1967 |
      | randomfile.flac    | 五月天 (Mayday)         | 自傳<autobiography>          | 擁抱/Embrace                  | 3     | Pop        | 2016 |
      | track.wav          | " ..Daft Punk.. "       | " Random Access Memories "   | " Get Lucky feat. Pharrell "  | 10    | Electronic | 2013 |
      | file.m4a           | Björk & Thom Yorke     | Medúlla Remixes*Extended     | Desired Constellation?        | 1     | Experimental | 2004 |
    And an empty target directory "target_complete_tags"
    When I run the command "muxic copy --source source_complete_tags --target target_complete_tags"
    Then the destination files should be created based entirely on tag metadata, not original filenames:
      | Expected Path |
      | target_complete_tags/The Beatles/Sgt. Pepper's Lonely Hearts Club Band/08 - Lucy In The Sky With Diamonds.mp3 |
      | target_complete_tags/Wu Yue Tian (Mayday)/Zi Chuan -autobiography-/03 - Yong Bao-Embrace.flac |
      | target_complete_tags/Daft Punk/Random Access Memories/10 - Get Lucky Ft. Pharrell.wav |
      | target_complete_tags/Bjork and Thom Yorke/Medullla Remixes-Extended/01 - Desired Constellation-.m4a |
    And all folder names should be Windows filesystem compatible (no prohibited characters, no leading/trailing periods or spaces)
    And all file names should be Windows filesystem compatible (no prohibited characters, no leading/trailing periods or spaces)
    And Unicode characters should be converted to closest ASCII equivalents
    And the console output should contain "Finished:" for each processed file

  Scenario: Tag-based filename generation handles missing metadata gracefully
    Given a source directory "source_missing_tags" with the following music files:
      | File Path           | Artist      | Album    | Title    | Track |
      | no_metadata.mp3     |             |          |          |       |
      | partial_tags.flac   | Known Artist|          | Known Title |   |
      | filename_only.wav   |             |          |          |       |
    And an empty target directory "target_missing_tags"
    When I run the command "muxic copy --source source_missing_tags --target target_missing_tags"
    Then files with missing metadata should use default values:
      | Expected Path |
      | target_missing_tags/Unknown/Unknown/01 - no_metadata.mp3 |
      | target_missing_tags/Known Artist/Unknown/01 - Known Title.flac |
      | target_missing_tags/Unknown/Unknown/01 - filename_only.wav |
    And the console output should contain "Warning: could not read tags" or similar metadata reading warnings
