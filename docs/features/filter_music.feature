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
