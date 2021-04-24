# muxic
I know what you're thinking -- why not just use "Beets"? 

Well, despite being vast and powerful, like a lot of open software Beets is presently broken for some dependencies and despite being maintained regularly by its volunteers, I actually found it to be super-complicated and overwrought for what I needed. I take a hard line on my music collection and I wanted a command-line tool that would simply sweep up the mess and not rely on me for all kinds of choices. So this tool does have some hard and fast rules at the moment. 

If you want flexibility and options, Beets is probably for you. If you want quick and decison-free, muxic is what you want. Feel free to contribute! 

# Usage

muxic [source folder of mess] [target folder for clean music]

Note that music will be *moved*, not *copied* from the source folder into the target folder. Where tags are not present and a filename can't be discerned, it will be copied into the root folder AS IS. Otherwise, expect folder layout in the following format:

[Artist]/[Album]/[Track Number] - [Track Title].mp3


