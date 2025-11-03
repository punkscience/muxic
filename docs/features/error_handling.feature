Feature: Error Handling in muxic CLI
  As a user of the muxic CLI,
  I want the application to handle errors gracefully,
  so that I understand what went wrong and the application doesn't crash unexpectedly.

  Scenario: Source directory does not exist
    Given a source directory "non_existent_source_dir" that does not exist
    And an empty target directory "target_dir_for_error"
    When I run the command "muxic copy --source non_existent_source_dir --target target_dir_for_error"
    Then the command should indicate failure or produce error output
    And the console output should contain "Error accessing path non_existent_source_dir" or "Error walking the path non_existent_source_dir"
    And the target directory "target_dir_for_error" should remain empty or not be created

  Scenario: Invalid source path provided (e.g., a file instead of a directory)
    Given a file named "source_is_a_file.txt" exists
    And an empty target directory "target_for_invalid_source"
    When I run the command "muxic copy --source source_is_a_file.txt --target target_for_invalid_source"
    Then the command should indicate failure or produce error output
    And the console output should contain an error message indicating the source is not a directory or cannot be read as such

  Scenario: Target directory is a file
    Given a source directory "source_for_file_target" with a music file "song.mp3"
    And a file named "target_is_a_file.txt" exists
    When I run the command "muxic copy --source source_for_file_target --target target_is_a_file.txt"
    Then the command should indicate failure or produce error output
    And the console output should contain "Error creating target folder" or a similar message indicating the target path is invalid

  Scenario: No write permission for the target directory
    Given a source directory "source_no_permission" with a music file "track.mp3"
    And a target directory "target_no_permission" where the user does not have write permissions
    When I run the command "muxic copy --source source_no_permission --target target_no_permission"
    Then the command should indicate failure or produce error output for each file
    And the console output should contain "Error creating target folder" or "Error copying file" along with a permission denied message

  Scenario: Corrupted music file (cannot read tags)
    Given a source directory "source_corrupted_file" with a "corrupted_song.mp3" that has invalid/unreadable tags
    And an empty target directory "target_for_corrupted"
    When I run the command "muxic copy --source source_corrupted_file --target target_for_corrupted"
    Then the file "target_for_corrupted/Unknown Artist/Unknown Album/01 - corrupted_song.mp3" should possibly be created (or handled by default naming)
    Or the console output should contain an error/warning about failing to read tags for "corrupted_song.mp3"
    # This depends on how movemusic library handles tag read failures. The key is it shouldn't crash.

  Scenario: Insufficient disk space in target directory
    Given a source directory "source_large_files" with a music file "very_large_song.mp3" (e.g., 1GB)
    And a target directory "target_low_space" with insufficient disk space for "very_large_song.mp3"
    When I run the command "muxic copy --source source_large_files --target target_low_space"
    Then the command should indicate failure for that file
    And the console output should contain "Error copying file" and possibly an "insufficient disk space" message
    And the partial file in "target_low_space" should ideally be cleaned up

  Scenario: Required arguments missing (no source)
    When I run the command "muxic copy --target some_target"
    Then the command should fail
    And the help text or an error message indicating "--source" is required should be displayed

  Scenario: Required arguments missing (no target)
    When I run the command "muxic copy --source some_source"
    Then the command should fail
    And the help text or an error message indicating "--target" is required should be displayed
