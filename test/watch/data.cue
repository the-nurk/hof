import "strings"

users: [n=string]: { name: n, NAME: strings.ToUpper(n) }
users: {
	aaah: role: "cat"
	alice: role: "user"
	bob: role: "user"
	cow: role: "mooer"
	darth: role: "evil"
	mary: role: "admin"
}