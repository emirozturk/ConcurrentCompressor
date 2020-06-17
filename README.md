# Concurrent Compressor  
Concurrent compressor is a concurrent n-gram based semi-static compression algorithm. It aims to compress big text files faster with it's concurrent architecture.

## Features
 - Uses n-grams as symbols
 - n could be between 1-8
 - Compress operation consists of two stages.
 - Dictionary based. Creates dictionary at first stage.
 - Creates 6 streams concurrently.

## Usage
To compress a file with selected ngram length:

```
CC -c fileName length
```

CC will create a ".ccf" file in the same directory as an output. 
To decompress this file:

```
CC -d fileName.ccf
```
