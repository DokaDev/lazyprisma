# Contributing

I love pull requests from everyone!

When contributing to this repository, please first discuss the change you wish
to make via [issue](https://github.com/dokadev/lazyprisma/issues) or any other
method with the owners of this repository before making a change.

If you've never written Go in your life, that's completely fine. Go is widely
considered an easy-to-learn language, and lazyprisma's codebase is modest in
size, so if you're looking for an open source project to gain dev experience,
you've come to the right place.

## Go

This project is written in Go. Go is an opinionated language with strict idioms,
but some of those idioms are a little extreme. Some things we do differently:

1. There is no shame in using `self` as a receiver name in a struct method. In fact we encourage it.
2. There is no shame in prefixing an interface with `I` instead of suffixing with `er` when there are several methods on the interface.
3. If a struct implements an interface, we make it explicit with something like:

```go
var _ MyInterface = &MyStruct{}
```

This makes the intent clearer and means that if we fail to satisfy the interface
we'll get an error in the file that needs fixing.


We use GitHub issues to track public bugs. Report a bug by
[opening a new issue](https://github.com/dokadev/lazyprisma/issues/new) -- it's
that easy!


## Any contributions you make will be under the MIT Software License

In short, when you submit code changes, your submissions are understood to be
under the same [MIT License](http://choosealicense.com/licenses/mit/) that
covers the project. Feel free to contact the maintainers if that's a concern.

## Improvements

If you can think of any way to improve these docs, let us know.
