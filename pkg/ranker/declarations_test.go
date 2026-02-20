// SPDX-License-Identifier: MIT

package ranker

import (
	"testing"
)

func TestIsDeclarationLine_Go(t *testing.T) {
	cases := []struct {
		line string
		want bool
	}{
		{"func Foo() {", true},
		{"func (r *Receiver) Method() {", true},
		{"func(r *Receiver) Method() {", true},
		{"type Foo struct {", true},
		{"type Foo interface {", true},
		{"var x = 1", true},
		{"var(", true},
		{"const Pi = 3.14", true},
		{"const(", true},
		// Not declarations
		{"return x", false},
		{"x := foo()", false},
		{"if err != nil {", false},
		{"fmt.Println(x)", false},
		{"// func Foo is a comment", false},
	}

	for _, tc := range cases {
		got := IsDeclarationLine([]byte(tc.line), "Go")
		if got != tc.want {
			t.Errorf("IsDeclarationLine(%q, Go) = %v, want %v", tc.line, got, tc.want)
		}
	}
}

func TestIsDeclarationLine_Python(t *testing.T) {
	cases := []struct {
		line string
		want bool
	}{
		{"def foo():", true},
		{"class Bar:", true},
		{"async def baz():", true},
		{"return x", false},
		{"x = foo()", false},
	}

	for _, tc := range cases {
		got := IsDeclarationLine([]byte(tc.line), "Python")
		if got != tc.want {
			t.Errorf("IsDeclarationLine(%q, Python) = %v, want %v", tc.line, got, tc.want)
		}
	}
}

func TestIsDeclarationLine_JavaScript(t *testing.T) {
	cases := []struct {
		line string
		want bool
	}{
		{"function foo() {", true},
		{"function(x) {", true},
		{"class Foo {", true},
		{"const x = 1", true},
		{"let x = 1", true},
		{"var x = 1", true},
		{"export function foo() {", true},
		{"export default function foo() {", true},
		{"export class Foo {", true},
		{"export const x = 1", true},
		{"export let x = 1", true},
		{"export default class Foo {", true},
		{"console.log(x)", false},
		{"return x", false},
	}

	for _, tc := range cases {
		got := IsDeclarationLine([]byte(tc.line), "JavaScript")
		if got != tc.want {
			t.Errorf("IsDeclarationLine(%q, JavaScript) = %v, want %v", tc.line, got, tc.want)
		}
	}
}

func TestIsDeclarationLine_TypeScript(t *testing.T) {
	cases := []struct {
		line string
		want bool
	}{
		{"function foo() {", true},
		{"interface Foo {", true},
		{"type Foo = string", true},
		{"enum Color {", true},
		{"export interface Foo {", true},
		{"export type Foo = string", true},
		{"export enum Color {", true},
		{"console.log(x)", false},
	}

	for _, tc := range cases {
		got := IsDeclarationLine([]byte(tc.line), "TypeScript")
		if got != tc.want {
			t.Errorf("IsDeclarationLine(%q, TypeScript) = %v, want %v", tc.line, got, tc.want)
		}
	}
}

func TestIsDeclarationLine_TSX(t *testing.T) {
	// TSX should behave identically to TypeScript
	got := IsDeclarationLine([]byte("interface Props {"), "TSX")
	if !got {
		t.Error("TSX should recognize TypeScript patterns")
	}
}

func TestIsDeclarationLine_Rust(t *testing.T) {
	cases := []struct {
		line string
		want bool
	}{
		{"fn main() {", true},
		{"pub fn foo() {", true},
		{"pub(crate) fn bar() {", true},
		{"struct Foo {", true},
		{"pub struct Foo {", true},
		{"enum Color {", true},
		{"pub enum Color {", true},
		{"trait Display {", true},
		{"pub trait Display {", true},
		{"impl Foo {", true},
		{"type Alias = String;", true},
		{"pub type Alias = String;", true},
		{"const MAX: i32 = 100;", true},
		{"pub const MAX: i32 = 100;", true},
		{"static X: i32 = 0;", true},
		{"pub static X: i32 = 0;", true},
		{"let x = 5;", false},
		{"println!(x);", false},
	}

	for _, tc := range cases {
		got := IsDeclarationLine([]byte(tc.line), "Rust")
		if got != tc.want {
			t.Errorf("IsDeclarationLine(%q, Rust) = %v, want %v", tc.line, got, tc.want)
		}
	}
}

func TestIsDeclarationLine_Java(t *testing.T) {
	cases := []struct {
		line string
		want bool
	}{
		{"public class Foo {", true},
		{"public interface Bar {", true},
		{"public enum Color {", true},
		{"public abstract class Base {", true},
		{"class Foo {", true},
		{"interface Bar {", true},
		{"enum Color {", true},
		{"abstract class Base {", true},
		{"System.out.println(x);", false},
	}

	for _, tc := range cases {
		got := IsDeclarationLine([]byte(tc.line), "Java")
		if got != tc.want {
			t.Errorf("IsDeclarationLine(%q, Java) = %v, want %v", tc.line, got, tc.want)
		}
	}
}

func TestIsDeclarationLine_C(t *testing.T) {
	cases := []struct {
		line string
		want bool
	}{
		{"#define MAX 100", true},
		{"typedef int myint;", true},
		{"struct foo {", true},
		{"enum color {", true},
		{"union data {", true},
		{"printf(x);", false},
	}

	for _, tc := range cases {
		got := IsDeclarationLine([]byte(tc.line), "C")
		if got != tc.want {
			t.Errorf("IsDeclarationLine(%q, C) = %v, want %v", tc.line, got, tc.want)
		}
	}
}

func TestIsDeclarationLine_CPP(t *testing.T) {
	cases := []struct {
		line string
		want bool
	}{
		// C patterns should work
		{"#define MAX 100", true},
		{"typedef int myint;", true},
		{"struct foo {", true},
		// C++ specific
		{"class Foo {", true},
		{"namespace bar {", true},
		{"template<typename T>", true},
		{"std::cout << x;", false},
	}

	for _, tc := range cases {
		got := IsDeclarationLine([]byte(tc.line), "C++")
		if got != tc.want {
			t.Errorf("IsDeclarationLine(%q, C++) = %v, want %v", tc.line, got, tc.want)
		}
	}
}

func TestIsDeclarationLine_CSharp(t *testing.T) {
	cases := []struct {
		line string
		want bool
	}{
		{"public class Foo {", true},
		{"public interface Bar {", true},
		{"public enum Color {", true},
		{"public struct Point {", true},
		{"internal class Baz {", true},
		{"Console.WriteLine(x);", false},
	}

	for _, tc := range cases {
		got := IsDeclarationLine([]byte(tc.line), "C#")
		if got != tc.want {
			t.Errorf("IsDeclarationLine(%q, C#) = %v, want %v", tc.line, got, tc.want)
		}
	}
}

func TestIsDeclarationLine_Ruby(t *testing.T) {
	cases := []struct {
		line string
		want bool
	}{
		{"def foo", true},
		{"class Bar", true},
		{"module Baz", true},
		{"puts x", false},
	}

	for _, tc := range cases {
		got := IsDeclarationLine([]byte(tc.line), "Ruby")
		if got != tc.want {
			t.Errorf("IsDeclarationLine(%q, Ruby) = %v, want %v", tc.line, got, tc.want)
		}
	}
}

func TestIsDeclarationLine_PHP(t *testing.T) {
	cases := []struct {
		line string
		want bool
	}{
		{"function foo() {", true},
		{"class Bar {", true},
		{"interface Baz {", true},
		{"trait Qux {", true},
		{"abstract class Base {", true},
		{"public function bar() {", true},
		{"private function baz() {", true},
		{"protected function qux() {", true},
		{"echo $x;", false},
	}

	for _, tc := range cases {
		got := IsDeclarationLine([]byte(tc.line), "PHP")
		if got != tc.want {
			t.Errorf("IsDeclarationLine(%q, PHP) = %v, want %v", tc.line, got, tc.want)
		}
	}
}

func TestIsDeclarationLine_Kotlin(t *testing.T) {
	cases := []struct {
		line string
		want bool
	}{
		{"fun main() {", true},
		{"class Foo {", true},
		{"data class Bar(val x: Int)", true},
		{"sealed class Base {", true},
		{"object Singleton {", true},
		{"interface Baz {", true},
		{"enum class Color {", true},
		{"typealias Name = String", true},
		{"val x = 1", true},
		{"var y = 2", true},
		{"println(x)", false},
	}

	for _, tc := range cases {
		got := IsDeclarationLine([]byte(tc.line), "Kotlin")
		if got != tc.want {
			t.Errorf("IsDeclarationLine(%q, Kotlin) = %v, want %v", tc.line, got, tc.want)
		}
	}
}

func TestIsDeclarationLine_Swift(t *testing.T) {
	cases := []struct {
		line string
		want bool
	}{
		{"func foo() {", true},
		{"class Bar {", true},
		{"struct Point {", true},
		{"enum Color {", true},
		{"protocol Displayable {", true},
		{"typealias Name = String", true},
		{"let x = 1", true},
		{"var y = 2", true},
		{"print(x)", false},
	}

	for _, tc := range cases {
		got := IsDeclarationLine([]byte(tc.line), "Swift")
		if got != tc.want {
			t.Errorf("IsDeclarationLine(%q, Swift) = %v, want %v", tc.line, got, tc.want)
		}
	}
}

func TestIsDeclarationLine_Shell(t *testing.T) {
	cases := []struct {
		line string
		want bool
	}{
		{"function my_func {", true},
		{"function(", true},
		{"echo hello", false},
		{"# function comment", false},
		{"my_func() {", false},
	}

	for _, tc := range cases {
		got := IsDeclarationLine([]byte(tc.line), "Shell")
		if got != tc.want {
			t.Errorf("IsDeclarationLine(%q, Shell) = %v, want %v", tc.line, got, tc.want)
		}
	}
}

func TestIsDeclarationLine_Lua(t *testing.T) {
	cases := []struct {
		line string
		want bool
	}{
		{"function foo()", true},
		{"local function bar()", true},
		{"print(x)", false},
		{"local x = 1", false},
	}

	for _, tc := range cases {
		got := IsDeclarationLine([]byte(tc.line), "Lua")
		if got != tc.want {
			t.Errorf("IsDeclarationLine(%q, Lua) = %v, want %v", tc.line, got, tc.want)
		}
	}
}

func TestIsDeclarationLine_Scala(t *testing.T) {
	cases := []struct {
		line string
		want bool
	}{
		{"def foo(): Unit = {", true},
		{"val x = 1", true},
		{"var y = 2", true},
		{"class Foo {", true},
		{"trait Bar {", true},
		{"object Baz {", true},
		{"case class Point(x: Int)", true},
		{"case object Nil", true},
		{"sealed trait Base", true},
		{"sealed class Node", true},
		{"abstract class Base", true},
		{"type Alias = String", true},
		{"println(x)", false},
		{"import scala.io", false},
	}

	for _, tc := range cases {
		got := IsDeclarationLine([]byte(tc.line), "Scala")
		if got != tc.want {
			t.Errorf("IsDeclarationLine(%q, Scala) = %v, want %v", tc.line, got, tc.want)
		}
	}
}

func TestIsDeclarationLine_Elixir(t *testing.T) {
	cases := []struct {
		line string
		want bool
	}{
		{"def foo do", true},
		{"defp bar do", true},
		{"defmodule MyApp do", true},
		{"defmacro my_macro do", true},
		{"defmacrop private_macro do", true},
		{"defstruct [:name, :age]", true},
		{"defprotocol Printable do", true},
		{"defimpl Printable, for: Atom do", true},
		{"IO.puts(x)", false},
		{":ok", false},
	}

	for _, tc := range cases {
		got := IsDeclarationLine([]byte(tc.line), "Elixir")
		if got != tc.want {
			t.Errorf("IsDeclarationLine(%q, Elixir) = %v, want %v", tc.line, got, tc.want)
		}
	}
}

func TestIsDeclarationLine_Haskell(t *testing.T) {
	cases := []struct {
		line string
		want bool
	}{
		{"data Color = Red | Green | Blue", true},
		{"type Name = String", true},
		{"newtype Wrapper a = Wrapper a", true},
		{"class Show a where", true},
		{"instance Show Color where", true},
		{"module Main where", true},
		{"import Data.List", false},
		{"main = putStrLn", false},
	}

	for _, tc := range cases {
		got := IsDeclarationLine([]byte(tc.line), "Haskell")
		if got != tc.want {
			t.Errorf("IsDeclarationLine(%q, Haskell) = %v, want %v", tc.line, got, tc.want)
		}
	}
}

func TestIsDeclarationLine_Perl(t *testing.T) {
	cases := []struct {
		line string
		want bool
	}{
		{"sub foo {", true},
		{"package MyModule;", true},
		{"use constant PI => 3.14;", true},
		{"print $x;", false},
		{"my $x = 1;", false},
	}

	for _, tc := range cases {
		got := IsDeclarationLine([]byte(tc.line), "Perl")
		if got != tc.want {
			t.Errorf("IsDeclarationLine(%q, Perl) = %v, want %v", tc.line, got, tc.want)
		}
	}
}

func TestIsDeclarationLine_Zig(t *testing.T) {
	cases := []struct {
		line string
		want bool
	}{
		{"fn main() void {", true},
		{"pub fn foo() void {", true},
		{"const x = 42;", true},
		{"pub const MAX = 100;", true},
		{"var y: i32 = 0;", true},
		{"pub var z: i32 = 0;", true},
		{"std.debug.print(x);", false},
		{"return x;", false},
	}

	for _, tc := range cases {
		got := IsDeclarationLine([]byte(tc.line), "Zig")
		if got != tc.want {
			t.Errorf("IsDeclarationLine(%q, Zig) = %v, want %v", tc.line, got, tc.want)
		}
	}
}

func TestIsDeclarationLine_Dart(t *testing.T) {
	cases := []struct {
		line string
		want bool
	}{
		{"class Foo {", true},
		{"abstract class Base {", true},
		{"enum Color {", true},
		{"mixin Printable {", true},
		{"extension StringExt on String {", true},
		{"typedef IntList = List<int>;", true},
		{"print(x);", false},
		{"var x = 1;", false},
	}

	for _, tc := range cases {
		got := IsDeclarationLine([]byte(tc.line), "Dart")
		if got != tc.want {
			t.Errorf("IsDeclarationLine(%q, Dart) = %v, want %v", tc.line, got, tc.want)
		}
	}
}

func TestIsDeclarationLine_Julia(t *testing.T) {
	cases := []struct {
		line string
		want bool
	}{
		{"function foo(x)", true},
		{"struct Point", true},
		{"mutable struct MPoint", true},
		{"abstract type Shape end", true},
		{"macro my_macro(ex)", true},
		{"module MyModule", true},
		{"println(x)", false},
		{"x = 1", false},
	}

	for _, tc := range cases {
		got := IsDeclarationLine([]byte(tc.line), "Julia")
		if got != tc.want {
			t.Errorf("IsDeclarationLine(%q, Julia) = %v, want %v", tc.line, got, tc.want)
		}
	}
}

func TestIsDeclarationLine_Clojure(t *testing.T) {
	cases := []struct {
		line string
		want bool
	}{
		{"(defn foo [x]", true},
		{"(def x 42)", true},
		{"(defmacro my-macro [x]", true},
		{"(defprotocol MyProto", true},
		{"(defrecord Point [x y])", true},
		{"(deftype MyType []", true},
		{"(ns my.namespace", true},
		{"(println x)", false},
		{"(+ 1 2)", false},
	}

	for _, tc := range cases {
		got := IsDeclarationLine([]byte(tc.line), "Clojure")
		if got != tc.want {
			t.Errorf("IsDeclarationLine(%q, Clojure) = %v, want %v", tc.line, got, tc.want)
		}
	}
}

func TestIsDeclarationLine_Erlang(t *testing.T) {
	cases := []struct {
		line string
		want bool
	}{
		{"-module(my_module).", true},
		{"-export([start/0]).", true},
		{"-define(MAX, 100).", true},
		{"-record(person, {name, age}).", true},
		{"-type color() :: red | green.", true},
		{"-spec foo(integer()) -> integer().", true},
		{"io:format(\"hello\").", false},
		{"X = 1.", false},
	}

	for _, tc := range cases {
		got := IsDeclarationLine([]byte(tc.line), "Erlang")
		if got != tc.want {
			t.Errorf("IsDeclarationLine(%q, Erlang) = %v, want %v", tc.line, got, tc.want)
		}
	}
}

func TestIsDeclarationLine_Groovy(t *testing.T) {
	cases := []struct {
		line string
		want bool
	}{
		{"def foo() {", true},
		{"class Foo {", true},
		{"interface Bar {", true},
		{"enum Color {", true},
		{"trait Printable {", true},
		{"println x", false},
		{"x = 1", false},
	}

	for _, tc := range cases {
		got := IsDeclarationLine([]byte(tc.line), "Groovy")
		if got != tc.want {
			t.Errorf("IsDeclarationLine(%q, Groovy) = %v, want %v", tc.line, got, tc.want)
		}
	}
}

func TestIsDeclarationLine_OCaml(t *testing.T) {
	cases := []struct {
		line string
		want bool
	}{
		{"let x = 42", true},
		{"type color = Red | Green", true},
		{"module M = struct", true},
		{"val x : int", true},
		{"external print : string -> unit", true},
		{"print_endline x", false},
		{"match x with", false},
	}

	for _, tc := range cases {
		got := IsDeclarationLine([]byte(tc.line), "OCaml")
		if got != tc.want {
			t.Errorf("IsDeclarationLine(%q, OCaml) = %v, want %v", tc.line, got, tc.want)
		}
	}
}

func TestIsDeclarationLine_MATLAB(t *testing.T) {
	cases := []struct {
		line string
		want bool
	}{
		{"function y = foo(x)", true},
		{"function [a, b] = bar(x)", true},
		{"disp(x)", false},
		{"x = 1;", false},
	}

	for _, tc := range cases {
		got := IsDeclarationLine([]byte(tc.line), "MATLAB")
		if got != tc.want {
			t.Errorf("IsDeclarationLine(%q, MATLAB) = %v, want %v", tc.line, got, tc.want)
		}
	}
}

func TestIsDeclarationLine_Powershell(t *testing.T) {
	cases := []struct {
		line string
		want bool
	}{
		{"function Get-Item {", true},
		{"class MyClass {", true},
		{"enum Color {", true},
		{"Write-Host $x", false},
		{"$x = 1", false},
	}

	for _, tc := range cases {
		got := IsDeclarationLine([]byte(tc.line), "Powershell")
		if got != tc.want {
			t.Errorf("IsDeclarationLine(%q, Powershell) = %v, want %v", tc.line, got, tc.want)
		}
	}
}

func TestIsDeclarationLine_Nim(t *testing.T) {
	cases := []struct {
		line string
		want bool
	}{
		{"proc foo(x: int): int =", true},
		{"func bar(): string =", true},
		{"type Color = enum", true},
		{"template myTemplate() =", true},
		{"macro myMacro() =", true},
		{"method draw(self: Shape) =", true},
		{"echo x", false},
		{"var x = 1", false},
	}

	for _, tc := range cases {
		got := IsDeclarationLine([]byte(tc.line), "Nim")
		if got != tc.want {
			t.Errorf("IsDeclarationLine(%q, Nim) = %v, want %v", tc.line, got, tc.want)
		}
	}
}

func TestIsDeclarationLine_Crystal(t *testing.T) {
	cases := []struct {
		line string
		want bool
	}{
		{"def foo", true},
		{"class Bar", true},
		{"module Baz", true},
		{"struct Point", true},
		{"enum Color", true},
		{"lib LibC", true},
		{"macro my_macro", true},
		{"puts x", false},
		{"x = 1", false},
	}

	for _, tc := range cases {
		got := IsDeclarationLine([]byte(tc.line), "Crystal")
		if got != tc.want {
			t.Errorf("IsDeclarationLine(%q, Crystal) = %v, want %v", tc.line, got, tc.want)
		}
	}
}

func TestIsDeclarationLine_V(t *testing.T) {
	cases := []struct {
		line string
		want bool
	}{
		{"fn main() {", true},
		{"pub fn foo() int {", true},
		{"struct Point {", true},
		{"pub struct Config {", true},
		{"enum Color {", true},
		{"type Callback = fn(int) int", true},
		{"const x = 42", true},
		{"println(x)", false},
		{"mut x := 1", false},
	}

	for _, tc := range cases {
		got := IsDeclarationLine([]byte(tc.line), "V")
		if got != tc.want {
			t.Errorf("IsDeclarationLine(%q, V) = %v, want %v", tc.line, got, tc.want)
		}
	}
}

func TestIsDeclarationLine_UnknownLanguage(t *testing.T) {
	got := IsDeclarationLine([]byte("func foo() {"), "BrainFuck")
	if got {
		t.Error("Unknown language should never return true")
	}
}

func TestHasDeclarationPatterns(t *testing.T) {
	known := []string{"Go", "Python", "JavaScript", "TypeScript", "TSX", "Rust", "Java", "C", "C++", "C#", "Ruby", "PHP", "Kotlin", "Swift", "Shell", "Lua", "Scala", "Elixir", "Haskell", "Perl", "Zig", "Dart", "Julia", "Clojure", "Erlang", "Groovy", "OCaml", "MATLAB", "Powershell", "Nim", "Crystal", "V"}
	for _, lang := range known {
		if !HasDeclarationPatterns(lang) {
			t.Errorf("HasDeclarationPatterns(%q) = false, want true", lang)
		}
	}

	if HasDeclarationPatterns("UnknownLang") {
		t.Error("HasDeclarationPatterns(UnknownLang) = true, want false")
	}
}

func TestClassifyMatchLocations_Go(t *testing.T) {
	content := []byte("func Foo() {\n\tx := Foo()\n}\n")
	// "Foo" appears at offset 5 (declaration line) and offset 19 (usage line)
	matchLocations := map[string][][]int{
		"Foo": {{5, 8}, {19, 22}},
	}

	decl, usage := ClassifyMatchLocations(content, matchLocations, "Go")

	if len(decl["Foo"]) != 1 {
		t.Fatalf("expected 1 declaration, got %d", len(decl["Foo"]))
	}
	if decl["Foo"][0][0] != 5 {
		t.Errorf("declaration at offset %d, want 5", decl["Foo"][0][0])
	}

	if len(usage["Foo"]) != 1 {
		t.Fatalf("expected 1 usage, got %d", len(usage["Foo"]))
	}
	if usage["Foo"][0][0] != 19 {
		t.Errorf("usage at offset %d, want 19", usage["Foo"][0][0])
	}
}

func TestClassifyMatchLocations_MultipleFunctions(t *testing.T) {
	content := []byte("func Foo() {\n}\nfunc Bar() {\n\tFoo()\n\tBar()\n}\n")
	// func Foo() {\n  -> line 0, offsets 0-13
	// }\n             -> line 1, offsets 14-15
	// func Bar() {\n  -> line 2, offsets 16-29
	// \tFoo()\n       -> line 3, offsets 30-36
	// \tBar()\n       -> line 4, offsets 37-43
	matchLocations := map[string][][]int{
		"Foo": {{5, 8}, {31, 34}},
		"Bar": {{21, 24}, {38, 41}},
	}

	decl, usage := ClassifyMatchLocations(content, matchLocations, "Go")

	if len(decl["Foo"]) != 1 || decl["Foo"][0][0] != 5 {
		t.Errorf("Foo declaration: got %v, want [{5 8}]", decl["Foo"])
	}
	if len(usage["Foo"]) != 1 || usage["Foo"][0][0] != 31 {
		t.Errorf("Foo usage: got %v, want [{31 34}]", usage["Foo"])
	}
	if len(decl["Bar"]) != 1 || decl["Bar"][0][0] != 21 {
		t.Errorf("Bar declaration: got %v, want [{21 24}]", decl["Bar"])
	}
	if len(usage["Bar"]) != 1 || usage["Bar"][0][0] != 38 {
		t.Errorf("Bar usage: got %v, want [{38 41}]", usage["Bar"])
	}
}

func TestClassifyMatchLocations_NoLanguage(t *testing.T) {
	content := []byte("func Foo() {\n}\n")
	matchLocations := map[string][][]int{
		"Foo": {{5, 8}},
	}

	decl, usage := ClassifyMatchLocations(content, matchLocations, "UnknownLang")

	if len(decl["Foo"]) != 0 {
		t.Errorf("unknown language should have no declarations, got %d", len(decl["Foo"]))
	}
	if len(usage["Foo"]) != 1 {
		t.Errorf("unknown language should classify all as usages, got %d", len(usage["Foo"]))
	}
}

func TestClassifyMatchLocations_Empty(t *testing.T) {
	// Empty content
	decl, usage := ClassifyMatchLocations([]byte{}, map[string][][]int{"x": {{0, 1}}}, "Go")
	if len(decl) != 0 || len(usage) != 0 {
		t.Error("empty content should return empty maps")
	}

	// Empty match locations
	decl, usage = ClassifyMatchLocations([]byte("func Foo()"), map[string][][]int{}, "Go")
	if len(decl) != 0 || len(usage) != 0 {
		t.Error("empty match locations should return empty maps")
	}

	// Nil match locations
	decl, usage = ClassifyMatchLocations([]byte("func Foo()"), nil, "Go")
	if len(decl) != 0 || len(usage) != 0 {
		t.Error("nil match locations should return empty maps")
	}
}

func TestClassifyMatchLocations_IndentedDeclaration(t *testing.T) {
	content := []byte("package main\n\n\tfunc Foo() {\n\t}\n")
	// "\tfunc Foo() {\n" starts at offset 14; "Foo" at offset 20
	matchLocations := map[string][][]int{
		"Foo": {{20, 23}},
	}

	decl, usage := ClassifyMatchLocations(content, matchLocations, "Go")

	if len(decl["Foo"]) != 1 {
		t.Errorf("indented declaration should be recognized, got declarations=%v usages=%v", decl["Foo"], usage["Foo"])
	}
}

func TestFindLine(t *testing.T) {
	// Simulates content: "abc\ndef\nghi\n"
	// Line 0: offset 0
	// Line 1: offset 4
	// Line 2: offset 8
	lineStarts := []int{0, 4, 8}

	cases := []struct {
		offset int
		want   int
	}{
		{0, 0},  // start of line 0
		{2, 0},  // middle of line 0
		{3, 0},  // end of line 0 (the \n)
		{4, 1},  // start of line 1
		{6, 1},  // middle of line 1
		{8, 2},  // start of line 2
		{10, 2}, // end of content
	}

	for _, tc := range cases {
		got := findLine(lineStarts, tc.offset)
		if got != tc.want {
			t.Errorf("findLine(%v, %d) = %d, want %d", lineStarts, tc.offset, got, tc.want)
		}
	}
}

func TestFindLine_SingleLine(t *testing.T) {
	lineStarts := []int{0}
	got := findLine(lineStarts, 5)
	if got != 0 {
		t.Errorf("findLine([0], 5) = %d, want 0", got)
	}
}

func TestClassifyMatchLocations_OutOfBounds(t *testing.T) {
	content := []byte("func Foo()\n")
	matchLocations := map[string][][]int{
		"Foo": {{-1, 2}, {100, 103}},
	}

	decl, usage := ClassifyMatchLocations(content, matchLocations, "Go")

	// Out of bounds matches should be classified as usages
	if len(decl["Foo"]) != 0 {
		t.Errorf("out of bounds should not be declarations, got %d", len(decl["Foo"]))
	}
	if len(usage["Foo"]) != 2 {
		t.Errorf("out of bounds should be usages, got %d", len(usage["Foo"]))
	}
}

func TestSupportedDeclarationLanguages(t *testing.T) {
	langs := SupportedDeclarationLanguages()
	if len(langs) < 32 {
		t.Errorf("expected at least 32 supported languages, got %d", len(langs))
	}

	// Check that known languages are present
	langSet := make(map[string]bool)
	for _, l := range langs {
		langSet[l] = true
	}
	required := []string{"Go", "Python", "JavaScript", "TypeScript", "TSX", "Rust", "Java", "C", "C++", "C#", "Ruby", "PHP", "Kotlin", "Swift", "Shell", "Lua", "Scala", "Elixir", "Haskell", "Perl", "Zig", "Dart", "Julia", "Clojure", "Erlang", "Groovy", "OCaml", "MATLAB", "Powershell", "Nim", "Crystal", "V"}
	for _, r := range required {
		if !langSet[r] {
			t.Errorf("missing required language: %s", r)
		}
	}
}
