# go2chatgpt
A small program that chunks up program files to be loaded into chatgpt. 1/2 written by chatgpt. Inspired by https://github.com/mpoon/gpt-repository-loader. Note that this has limitations as chatgpt doesn't seem to remember past 32KB-ish of data. This is probably documented somewhere but I haven't looked it up.

Ideally you could load up all the `chunk{num}.txt`s and query chatgpt questions in plain english as a tool to learn a new codebase.

## Install
```
git clone https://github.com/zricethezav/go2chatgpt.git
cd go2chatgpt
go build (install go dependencies)
```

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

## Examples
See `/example`


