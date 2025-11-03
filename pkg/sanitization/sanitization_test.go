package sanitization

import (
	"testing"
)

func TestWindowsSanitizer_SanitizeForFilesystem(t *testing.T) {
	sanitizer := NewWindowsSanitizer()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		// Basic cases
		{"empty string", "", ""},
		{"simple text", "Hello World", "Hello World"},
		{"already clean", "Artist Name", "Artist Name"},

		// Windows prohibited characters - based on copy_music.feature scenarios
		{"forward slash", "AC/DC", "AC-DC"},
		{"backslash", "Back\\Black", "Back-Black"},
		{"colon", "Album:Vol1", "Album-Vol1"},
		{"asterisk", "Hells*Bells", "Hells-Bells"},
		{"question mark", "Album?Mix", "Album-Mix"},
		{"double quotes", "Song\"Title\"", "Song-Title-"},
		{"angle brackets", "Artist<Name>", "Artist-Name-"},
		{"pipe", "Band|Name", "Band-Name"},
		{"multiple prohibited chars", "AC/DC\\Back*Black?", "AC-DC-Back-Black-"},

		// Complex prohibited character combinations from feature files
		{"complex artist", "Artist<Name>", "Artist-Name-"},
		{"complex album", "Album?Mix", "Album-Mix"},
		{"complex title", "Song\"Title\"", "Song-Title-"},
		{"band with pipe", "Band|Name", "Band-Name"},
		{"track with greater than", "Track>Name", "Track-Name"},

		// Unicode and non-ASCII characters - based on copy_music.feature scenarios
		{"icelandic characters", "Björk", "Bjork"},
		{"icelandic album", "Médulla", "Medulla"},
		{"icelandic title", "Öll Birtan", "Oll Birtan"},
		{"icelandic band", "Sigur Rós", "Sigur Ros"},
		{"icelandic album 2", "Ágætis byrjun", "Agaetis Byrjun"},
		{"icelandic title 2", "Svefn-g-englar", "Svefn-G-Englar"},
		{"chinese artist", "中文歌手", "Zhong Wen Ge Shou"},
		{"chinese album", "专辑名称", "Zhuan Ji Ming Cheng"},
		{"chinese title", "歌曲标题", "Ge Qu Biao Ti"},
		{"japanese text", "音楽アルバム", "Yin Le Arubamu"},
		{"cafe accent", "Café Del Mar", "Cafe Del Mar"},
		{"naive accent", "Naïve Song", "Naive Song"},

		// Leading and trailing spaces and periods - based on copy_music.feature scenarios
		{"leading trailing spaces", " Artist ", "Artist"},
		{"leading trailing spaces album", " Album Name ", "Album Name"},
		{"leading trailing spaces title", " Song Title ", "Song Title"},
		{"leading trailing periods", "..Artist..", "Artist"},
		{"leading trailing periods album", "..Album.Name..", "Album.Name"},
		{"leading trailing periods title", "..Song.Title..", "Song.Title"},
		{"mixed spaces and periods", ". Artist .", "Artist"},
		{"mixed spaces and periods album", ". Album .", "Album"},
		{"mixed spaces and periods title", ". Title .", "Title"},

		// Complex combinations from feature files
		{"motley crue complex", " //Mötley\\Crüe// ", "--Motley-Crue--"},
		{"shout at devil complex", " ..Shout*At?The<Devil>.. ", "Shout-At-The-Devil-"},
		{"girls girls girls complex", " ..Girls,\"Girls\",Girls.. ", "Girls,-Girls-,Girls"},
		{"rock star complex", " //Rock\\Star// ", "--Rock-Star--"},
		{"greatest hits complex", " ..Greatest*Hits.. ", "Greatest-Hits"},
		{"hit song complex", " ..Hit<Song>Name.. ", "Hit-Song-Name"},

		// Substitutions - feat./featuring/& replacements
		{"feat period", "Artist feat. Other", "Artist Ft Other"},
		{"Feat capital", "Artist Feat. Other", "Artist Ft Other"},
		{"Feat no period", "Artist Feat Other", "Artist Ft Other"},
		{"featuring full", "Artist Featuring Other", "Artist Ft Other"},
		{"ampersand", "Artist & Band", "Artist And Band"},
		{"get lucky ft", "Get Lucky feat. Pharrell", "Get Lucky Ft Pharrell"},

		// Multiple spaces normalization
		{"double spaces", "Too  Much   Space", "Too Much Space"},
		{"multiple spaces mixed", "Artist  feat.   Other", "Artist Ft Other"},

		// Title casing
		{"lowercase", "lowercase title", "Lowercase Title"},
		{"mixed case", "MiXeD cAsE", "Mixed Case"},
		{"all caps", "ALL CAPS TITLE", "All Caps Title"},

		// Edge cases from feature files
		{"parentheses album", "( )", "( )"},
		{"untitled track", "Untitled #1", "Untitled #1"},
		{"embrace slash", "擁抱/Embrace", "Yong Bao-Embrace"},
		{"mayday chinese", "五月天 (Mayday)", "Wu Yue Tian (Mayday)"},
		{"autobiography brackets", "自傳<autobiography>", "Zi Chuan -Autobiography-"},
		{"bjork collaboration", "Björk & Thom Yorke", "Bjork And Thom Yorke"},
		{"medulla extended", "Medúlla Remixes*Extended", "Medulla Remixes-Extended"},
		{"desired constellation", "Desired Constellation?", "Desired Constellation-"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := sanitizer.SanitizeForFilesystem(tc.input)
			if result != tc.expected {
				t.Errorf("SanitizeForFilesystem(%q) = %q, expected %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestWindowsSanitizer_SanitizeFolderName(t *testing.T) {
	sanitizer := NewWindowsSanitizer()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple folder", "Music Folder", "Music Folder"},
		{"folder with slash", "Rock/Metal", "Rock-Metal"},
		{"folder with periods", "..Folder..", "Folder"},
		{"folder with spaces", " Folder Name ", "Folder Name"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := sanitizer.SanitizeFolderName(tc.input)
			if result != tc.expected {
				t.Errorf("SanitizeFolderName(%q) = %q, expected %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestWindowsSanitizer_SanitizeFileName(t *testing.T) {
	sanitizer := NewWindowsSanitizer()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple file", "song.mp3", "Song.mp3"},
		{"file with prohibited chars", "song*.mp3", "Song-.mp3"},
		{"file with unicode", "sång.mp3", "Sang.mp3"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := sanitizer.SanitizeFileName(tc.input)
			if result != tc.expected {
				t.Errorf("SanitizeFileName(%q) = %q, expected %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestWindowsSanitizer_SanitizeTrackMetadata(t *testing.T) {
	sanitizer := NewWindowsSanitizer()

	testCases := []struct {
		name           string
		artist         string
		album          string
		title          string
		expectedArtist string
		expectedAlbum  string
		expectedTitle  string
	}{
		{
			name:           "basic metadata",
			artist:         "The Beatles",
			album:          "Abbey Road",
			title:          "Come Together",
			expectedArtist: "The Beatles",
			expectedAlbum:  "Abbey Road",
			expectedTitle:  "Come Together",
		},
		{
			name:           "metadata with prohibited characters",
			artist:         "AC/DC",
			album:          "Back\\Black",
			title:          "Hells*Bells",
			expectedArtist: "AC-DC",
			expectedAlbum:  "Back-Black",
			expectedTitle:  "Hells-Bells",
		},
		{
			name:           "metadata with unicode",
			artist:         "Björk",
			album:          "Médulla",
			title:          "Öll Birtan",
			expectedArtist: "Bjork",
			expectedAlbum:  "Medulla",
			expectedTitle:  "Oll Birtan",      
		},
		{
			name:           "metadata with spaces and periods",
			artist:         " ..Artist.. ",
			album:          " ..Album.. ",
			title:          " ..Title.. ",
			expectedArtist: "Artist",
			expectedAlbum:  "Album",
			expectedTitle:  "Title",
		},
		{
			name:           "complex real-world example",
			artist:         " //Mötley\\Crüe// ",
			album:          " ..Shout*At?The<Devil>.. ",
			title:          " ..Girls,\"Girls\",Girls.. ",
			expectedArtist: "--Motley-Crue--",
			expectedAlbum:  "Shout-At-The-Devil-",
			expectedTitle:  "Girls,-Girls-,Girls",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			artist, album, title := sanitizer.SanitizeTrackMetadata(tc.artist, tc.album, tc.title)
			
			if artist != tc.expectedArtist {
				t.Errorf("SanitizeTrackMetadata artist: got %q, expected %q", artist, tc.expectedArtist)
			}
			if album != tc.expectedAlbum {
				t.Errorf("SanitizeTrackMetadata album: got %q, expected %q", album, tc.expectedAlbum)
			}
			if title != tc.expectedTitle {
				t.Errorf("SanitizeTrackMetadata title: got %q, expected %q", title, tc.expectedTitle)
			}
		})
	}
}

func TestValidateWindowsPath(t *testing.T) {
	testCases := []struct {
		name     string
		path     string
		expected bool
	}{
		{"empty path", "", false},
		{"valid path", "Music/Artist/Album", true},
		{"path with prohibited char", "Music/Art<ist/Album", false},
		{"path with question mark", "Music/Album?/Song", false},
		{"path with asterisk", "Music/Album*/Song", false},
		{"path with pipe", "Music/Artist|Band/Album", false},
		{"path with quotes", "Music/\"Artist\"/Album", false},
		{"path ending with period", "Music/Artist/Album.", false},
		{"path ending with space", "Music/Artist/Album ", false},
		{"path starting with period", "Music/.Artist/Album", false},
		{"path starting with space", "Music/ Artist/Album", false},
		{"valid complex path", "Music/The Beatles/Abbey Road", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ValidateWindowsPath(tc.path)
			if result != tc.expected {
				t.Errorf("ValidateWindowsPath(%q) = %t, expected %t", tc.path, result, tc.expected)
			}
		})
	}
}

func TestWindowsSanitizerWithCustomSubstitutions(t *testing.T) {
	customSubs := map[string]string{
		"vs.": "versus",
		"w/":  "with", 
		"@":   "at",
	}
	
	sanitizer := NewWindowsSanitizerWithSubstitutions(customSubs)

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"custom vs", "Artist vs. Other", "Artist Versus Other"},
		{"custom w/", "Song w/ Feature", "Song With Feature"},
		{"custom @", "Live @ Venue", "Live At Venue"},
		{"standard feat should not work", "Artist feat. Other", "Artist Feat. Other"}, // Should not be replaced since we used custom subs
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := sanitizer.SanitizeForFilesystem(tc.input)
			if result != tc.expected {
				t.Errorf("Custom sanitizer SanitizeForFilesystem(%q) = %q, expected %q", tc.input, result, tc.expected)
			}
		})
	}
}

// Benchmark tests to ensure performance is acceptable
func BenchmarkSanitizeForFilesystem(b *testing.B) {
	sanitizer := NewWindowsSanitizer()
	input := " //Mötley\\Crüe// feat. Other Artist & The Band"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sanitizer.SanitizeForFilesystem(input)
	}
}

func BenchmarkSanitizeTrackMetadata(b *testing.B) {
	sanitizer := NewWindowsSanitizer()
	artist := " //AC\\DC// "
	album := " ..Back*In?Black<>.. "
	title := " ..Thunder\"Strike|Rock.. "
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sanitizer.SanitizeTrackMetadata(artist, album, title)
	}
}

// Test edge cases and error conditions
func TestEdgeCases(t *testing.T) {
	sanitizer := NewWindowsSanitizer()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"only prohibited chars", "<>:\"|?*\\/", "---------"},
		{"only periods", ".....", ""},
		{"only spaces", "     ", ""},
		{"mixed periods and spaces", " . . . ", ""},
		{"single character", "a", "A"},
		{"unicode only", "中文", "Zhong Wen"},
		{"very long string", "This is a very long string that contains many words and should be processed correctly even though it is quite lengthy and has various characters including spaces and punctuation marks.", "This Is A Very Long String That Contains Many Words And Should Be Processed Correctly Even Though It Is Quite Lengthy And Has Various Characters Including Spaces And Punctuation Marks"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := sanitizer.SanitizeForFilesystem(tc.input)
			if result != tc.expected {
				t.Errorf("SanitizeForFilesystem(%q) = %q, expected %q", tc.input, result, tc.expected)
			}
		})
	}
}