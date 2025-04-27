package types

type Config struct {
  home string
  privateKey string
  publicKey string
}

func makeUser(user string) Config {

var home = "/home/" + user + ""

var privateKey = "" + home + "/.ssh/id_ed25519"

var publicKey = "" + privateKey + ".pub"

var config = Config{
  home: home,
  privateKey: privateKey,
  publicKey: publicKey,
}
  return config
}

var configs = []Config{
  makeUser("bill"),
  makeUser("jane"),
}

var Configs = configs
