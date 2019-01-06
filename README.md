# Rocky Arena of Scissors and Papers
Welcome to the Arena, Gladiator. This battlefield is where you can challenge
your peers to some Rock Paper Scissors matches.

Choose a move, a peer and a bet and send the challenge. If you see a challenge
in your pending section, you can hover over it and choose to accept it with a
move.

The Peerster blockchain takes care of ensuring you will not be cheated and can
show your might to the whole network.

To test an implementation with 4 players, run the script `./rasp.sh`. Then open
`localhost:3000/${PORT}` (PORT in {8000, 8001, 8002, 8003}) in your browser, or
or curl it, but you'll be missing the GUI :p

# Requirements:
To be able to run the game you need the following dependencies.
- [`yarn`](https://yarnpkg.com/en/docs/install#mac-stable) package manager
- golang go1.11.1

Once Yarn is present on your machine, don't forget to install all the required
js packages by running:
```
(cd www/; yarn install)
```

# Collaborators:
- [Remi Coudert](https://github.com/korf74)
- [Lucas Gauchoux](https://github.com/lggoch)

# Areas with possible improvements:
- I'm still learning React so am very open to feedback.
- It would be cool to have a dockerfile that facilitates running the game
- For game improvements please refer to this [report]('./RASP.pdf')
- Ability to block malicious players
- Ability to create Best Of ${$N+1} games
- Cache ledgers at common fork points.