# Peerster
Homework for Decentralized Systems Engineering at EPFL.

# Requirements:
- [`yarn`](https://yarnpkg.com/en/docs/install#mac-stable) package manager
- golang go1.11.1
Don't forget to install all the required js packages
```
(cd www/; yarn install)
```

# GUI Usage:
To make a file available for sharing, click on `Share File...` and select it,
this will copy it to `_SharedFiles` and automatically send a message containing
the name of the file and its metahash to the selected chat.

To download a file, enter the hexadecimal metahash and the name to give the
resulting file then press download.

# How to test:
```
sh hw2.sh
// go to localhost:3000
// This will create 4 gossipers (A, B, C, D), and B is in the subdirectory: test
// you can check its files in test/_SharedFiles and test/_Downloads
```
