// STDLIB

let map = fn(arr, f) {
    let iter = fn(arr, acc) {
        if (len(arr) == 0) {
            return acc;
        } else {
            return iter(rest(arr), push(acc, f(first(arr))));
        }
    };
    return iter(arr, []);
};

let reduce = fn(arr, initial, f) {
    let iter = fn(arr, result) {
        if (len(arr) == 0) {
            return result;
        } else {
            iter(rest(arr), f(result, first(arr)));
        }
    };
    return iter(arr, initial);
}

// TEST

let x = [1, 2, 3, 4];
println(x);
let xx = map(x, fn(x) { return x * 2; });
println(xx);
let xx = reduce(x, 0, fn(acc, n) { return acc + n; });
println(xx);
