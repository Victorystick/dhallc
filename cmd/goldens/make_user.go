package make_user

func makeUser(user string) struct {
  home string
  privateKey string
  publicKey string
} {

var home = "/home/" + user + ""

var privateKey = "" + home + "/.ssh/id_ed25519"

var publicKey = "" + privateKey + ".pub"
  return struct {
  home string
  privateKey string
  publicKey string
}{
  home: home,
  privateKey: privateKey,
  publicKey: publicKey,
}
}

var Bill = makeUser("bill")
var Jane = makeUser("jane")
