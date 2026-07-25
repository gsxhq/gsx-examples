package views

import "github.com/gsxhq/gsx"

// Patch replaces the contents of the marker region named `name`, wherever it
// already sits in the document — which is what lets a late panel fill an early
// hole. `for` is an ordinary HTML attribute here, not gsx control flow.
component Patch(name string, children gsx.Node) {
	<template for={name}>{ children }</template>
}
