package walnut

import (
	"fmt"
	"strings"
)

type item interface {
	fmt.Stringer
	Item()
}

type add struct {
	n int
}

func (a add) Item() {
}

func (a add) String() string {
	return strings.Repeat("+", a.N())
}

func (a add) N() int {
	return a.n
}

type sub struct {
	n int
}

func (s sub) Item() {
}

func (s sub) String() string {
	return strings.Repeat("-", s.N())
}

func (s sub) N() int {
	return s.n
}

type next struct {
	n int
}

func (n next) Item() {
}

func (n next) String() string {
	return strings.Repeat(">", n.N())
}

func (n next) N() int {
	return n.n
}

type prev struct {
	n int
}

func (p prev) Item() {
}

func (p prev) String() string {
	return strings.Repeat("<", p.N())
}

func (p prev) N() int {
	return p.n
}

type read struct{}

func (r read) Item() {
}

func (r read) String() string {
	return ","
}

type write struct{}

func (w write) Item() {
}

func (w write) String() string {
	return "."
}

type loopStart struct {
}

func (l loopStart) Item() {
}

func (l loopStart) String() string {
	return "["
}

type loopEnd struct {
}

func (l loopEnd) Item() {
}

func (l loopEnd) String() string {
	return "]"
}

type clear struct {
}

func (c clear) Item() {
}

func (c clear) String() string {
	return "[-]"
}

type parseError struct {
	msg string
}

func (p parseError) Item() {
}

func (p parseError) String() string {
	return p.msg
}

func (p parseError) Error() string {
	return p.msg
}
