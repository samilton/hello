using Go = import "go.capnp";
using Lua = import "lua.capnp";

$Go.package("capnp");
$Go.import("capnp");

@0xef0924d23362a047;

struct Event {
	timestamp @0 :Int64;
	class @1 :Text;
	message @2 :Text;
	source @3 :Text;
	origin @4 :Text;
}

