package ctxware

// Copyright 2016 Nick Stogner. All rights reserved.
// Use of this source code is governed by the MIT
// license which can be found in the LICENSE file.

/*
Package ctxware provides patterns for chaining http middleware that relies on
net/context. Middleware can depend on other middleware. The composition
functions check for these dependencies when they are called.
*/
