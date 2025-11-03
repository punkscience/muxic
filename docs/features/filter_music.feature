Feature: Filter Music Files
  As a user of the muxic CLI,
  I want to be able to filter music files during the copy process,
  so that I only process the files I'm interested in.

  Scenario: Filter by name (case-insensitive)
    Given a source directory "source_filter_name" with the following music files:
      | File Path        |
      | song_rock.mp3    |
      | pop_song.mp3     |
      | my_ROCK_ballad.flac |
    And an empty target directory "target_filter_name"
    When I run the command "muxic copy --source source_filter_name --target target_filter_name --filter ROCK"
    Then the file "target_filter_name/Artist/Album/01 - song_rock.mp3" should exist
    And the file "target_filter_name/Artist/Album/01 - my_ROCK_ballad.flac" should exist
    And the file "target_filter_name/Artist/Album/01 - pop_song.mp3" should not exist
    And the console output should contain "Filtering files to those containing: ROCK"

  Scenario: Filter by name with no matches
    Given a source directory "source_filter_no_match" with music files "song1.mp3, song2.mp3"
    And an empty target directory "target_filter_no_match"
    When I run the command "muxic copy --source source_filter_no_match --target target_filter_no_match --filter NonExistent"
    Then the target directory "target_filter_no_match" should remain empty or not be created if it didn't exist (excluding base folder)
    And the console output should contain "Found 0 music files"

  Scenario: Filter by size (over X MB)
    Given a source directory "source_filter_size" with the following music files:
      | File Path      | Size (MB) |
      | large_file.mp3 | 10        |
      | small_file.mp3 | 2         |
      | medium_file.wav| 5         |
    And an empty target directory "target_filter_size"
    When I run the command "muxic copy --source source_filter_size --target target_filter_size --over 5"
    Then the file "target_filter_size/Artist/Album/01 - large_file.mp3" should exist  # Assuming size 10MB
    And the file "target_filter_size/Artist/Album/01 - small_file.mp3" should not exist # Assuming size 2MB
    And the file "target_filter_size/Artist/Album/01 - medium_file.wav" should not exist # Assuming size 5MB (not strictly 'over 5')
    And the console output should contain "Filtering files to those containing:  and size > 5MB"

  Scenario: Filter by size (over X MB) where X is large and no files match
    Given a source directory "source_filter_size_large" with music files "file1.mp3 (5MB), file2.flac (8MB)"
    And an empty target directory "target_filter_size_large"
    When I run the command "muxic copy --source source_filter_size_large --target target_filter_size_large --over 10"
    Then the target directory "target_filter_size_large" should remain empty
    And the console output should contain "Found 0 music files"

  Scenario: Filter by name AND size
    Given a source directory "source_filter_combo" with the following music files:
      | File Path         | Size (MB) |
      | project_alpha.mp3 | 12        |
      | project_beta.mp3  | 3         |
      | another_song.flac | 15        |
    And an empty target directory "target_filter_combo"
    When I run the command "muxic copy --source source_filter_combo --target target_filter_combo --filter project --over 10"
    Then the file "target_filter_combo/Artist/Album/01 - project_alpha.mp3" should exist
    And the file "target_filter_combo/Artist/Album/01 - project_beta.mp3" should not exist
    And the file "target_filter_combo/Artist/Album/01 - another_song.flac" should not exist
    And the console output should contain "Filtering files to those containing: project and size > 10MB"

  Scenario: Filter by name (case-insensitive) and move
    Given a source directory "source_filter_move" with files "filter_me.mp3", "dont_filter.mp3"
    And an empty target directory "target_filter_move"
    When I run the command "muxic copy --source source_filter_move --target target_filter_move --filter ME --move"
    Then the file "target_filter_move/Artist/Album/01 - filter_me.mp3" should exist
    And the source file "source_filter_move/filter_me.mp3" should not exist
    And the source file "source_filter_move/dont_filter.mp3" should still exist

  Scenario: Filter with Windows filesystem sanitization - prohibited characters in metadata
    Given a source directory "source_filter_sanitize" with the following music files:
      | File Path           | Artist        | Album          | Title             |
      | rock_song.mp3       | AC/DC         | Back\\Black    | Hells*Bells      |
      | pop_song.flac       | Taylor/Swift  | 1989<Deluxe>   | Shake?It"Off     |
      | jazz_track.wav      | Miles|Davis   | Kind:Of<Blue>  | So*What          |
    And an empty target directory "target_filter_sanitize"
    When I run the command "muxic copy --source source_filter_sanitize --target target_filter_sanitize --filter rock"
    Then the file "target_filter_sanitize/AC-DC/Back-Black/01 - Hells-Bells.mp3" should exist
    And the file "target_filter_sanitize/Taylor-Swift/1989-Deluxe-/01 - Shake-It-Off.flac" should not exist
    And the file "target_filter_sanitize/Miles-Davis/Kind-Of-Blue-/01 - So-What.wav" should not exist
    And the console output should contain "Filtering files to those containing: rock"

  Scenario: Filter with Unicode character sanitization  
    Given a source directory "source_filter_unicode" with the following music files:
      | File Path         | Artist      | Album         | Title           |
      | björk_song.mp3    | Björk       | Médulla       | Öll Birtan     |
      | sigur_song.flac   | Sigur Rós   | Ágætis byrjun | Svefn-g-englar |
      | chinese_song.wav  | 中文歌手     | 专辑名称       | 歌曲标题        |
    And an empty target directory "target_filter_unicode"
    When I run the command "muxic copy --source source_filter_unicode --target target_filter_unicode --filter björk"
    Then the file "target_filter_unicode/Bjork/Medulla/01 - Oll Birtan.mp3" should exist
    And the file "target_filter_unicode/Sigur Ros/Agaetis Byrjun/01 - Svefn-G-Englar.flac" should not exist
    And the file "target_filter_unicode/Zhong Wen Ge Shou/Zhuan Ji Ming Cheng/01 - Ge Qu Biao Ti.wav" should not exist

  Scenario: Filter with leading/trailing spaces and periods sanitization
    Given a source directory "source_filter_trim" with the following music files:
      | File Path        | Artist         | Album           | Title             |
      | spaced_song.mp3  | " Rock Band "  | " Rock Album "  | " Rock Song "    |
      | period_song.flac | "..Rock.Art.." | "..Rock.LP.."   | "..Rock.Track.." |
      | mixed_song.wav   | ". Jazz . "    | ". Jazz LP . "  | ". Jazz Tune . " |
    And an empty target directory "target_filter_trim"
    When I run the command "muxic copy --source source_filter_trim --target target_filter_trim --filter rock"
    Then the file "target_filter_trim/Rock Band/Rock Album/01 - Rock Song.mp3" should exist
    And the file "target_filter_trim/Rock.Art/Rock.LP/01 - Rock.Track.flac" should exist
    And the file "target_filter_trim/Jazz/Jazz LP/01 - Jazz Tune.wav" should not exist

  Scenario: Filter with complex sanitization and size constraints
    Given a source directory "source_filter_complex" with the following music files:
      | File Path         | Artist               | Album                  | Title                    | Size (MB) |
      | big_messy.mp3     | " //Rock\\Star// "   | " ..Greatest*Hits.. "  | " ..Hit<Song>Name.. "   | 12        |
      | small_clean.flac  | Clean Artist         | Clean Album            | Clean Title              | 3         |
      | big_other.wav     | Different Band       | Other Album            | Other Song               | 15        |
    And an empty target directory "target_filter_complex"  
    When I run the command "muxic copy --source source_filter_complex --target target_filter_complex --filter rock --over 10"
    Then the file "target_filter_complex/--Rock-Star--/Greatest-Hits/01 - Hit-Song-Name.mp3" should exist
    And the file "target_filter_complex/Clean Artist/Clean Album/01 - Clean Title.flac" should not exist
    And the file "target_filter_complex/Different Band/Other Album/01 - Other Song.wav" should not exist
    And the console output should contain "Filtering files to those containing: rock and size > 10MB"
