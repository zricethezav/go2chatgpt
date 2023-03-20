# go2chatgpt
A small program that chunks up program files to be loaded into chatgpt. 1/2 written by chatgpt. Inspired by https://github.com/mpoon/gpt-repository-loader

## Usage
```
Usage: go2chatgpt [options] <source> <output_folder>
  -chunksize int
        Chunk size in KB (GPT-4 max is 16KB) (default 13)
  -exclude string
        Comma-separated list of glob patterns to exclude
  -include string
        Comma-separated list of glob patterns to include
```


