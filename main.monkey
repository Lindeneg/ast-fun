// check variable block-scopes

// start by declaring two variables
let first = "Foo";
let last = "Bar";

// make a function that takes
// parameters with the same names
let make_name = fn(first, last) {
    return first + " " + last;
};

let first_ = "Christian";
let last_ = "Lindeneg";

// should be Christian Lindeneg and not Foo Bar
let name = make_name(first_, last_)

// lets see if it worked
println(name);

let x = 10;
let y = 20;

println(x, "+", y, "=", x + y);

