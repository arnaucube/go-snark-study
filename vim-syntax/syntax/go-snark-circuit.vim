" Vim syntax file
" Language:	go-snark-circuit
" URL:		https://github.com/arnaucube/go-snark/blob/master/vim-syntax/syntax/go-snark-circuit.vim

if !exists("main_syntax")
  " quit when a syntax file was already loaded
  if exists("b:current_syntax")
    finish
  endif
  let main_syntax = 'go-snark-circuit'
elseif exists("b:current_syntax") && b:current_syntax == "go-snark-circuit"
  finish
endif

let s:cpo_save = &cpo
set cpo&vim

syn keyword goSnarkCircuitCommentTodo      TODO FIXME XXX TBD contained
syn match   goSnarkCircuitLineComment      "\/\/.*" contains=@Spell,goSnarkCircuitCommentTodo
syn match   goSnarkCircuitSpecialCharacter "'\\.'"
syn match   goSnarkCircuitNumber	       "-\=\<\d\+L\=\>\|0[xX][0-9a-fA-F]\+\>"
syn match goSnarkCircuitOpSymbols "+\|-\|\*\|:\|)\|(\|="
syn keyword goSnarkCircuitPrivatePublic		private public
syn keyword goSnarkCircuitOut	out
syn keyword goSnarkCircuitEquals	equals
syn keyword goSnarkCircuitFunction	func
syn match goSnarkCircuitFuncCall /\<\K\k*\ze\s*(/
syn keyword goSnarkCircuitPrivate private nextgroup=goSnarkCircuitInputName skipwhite
syn keyword goSnarkCircuitPublic public nextgroup=goSnarkCircuitInputName skipwhite
syn match goSnarkCircuitInputName '\i\+' contained
syn match	goSnarkCircuitBraces	   "[{}\[\]]"
syn match	goSnarkCircuitParens	   "[()]"

syn sync fromstart
syn sync maxlines=100

" Define the default highlighting.
" Only when an item doesn't have highlighting yet
hi def link goSnarkCircuitLineComment		Comment
hi def link goSnarkCircuitCommentTodo		Todo
hi def link goSnarkCircuitSpecialCharacter	Special
hi def link goSnarkCircuitNumber		Number
hi def link goSnarkCircuitOpSymbols		Operator
hi def link goSnarkCircuitFuncCall		Function
hi def link goSnarkCircuitEquals		Identifier
hi def link goSnarkCircuitFunction		Keyword
hi def link goSnarkCircuitBraces		Function
hi def link goSnarkCircuitPrivate 		Keyword
hi def link goSnarkCircuitPublic		Keyword
hi def link goSnarkCircuitInputName		Special
hi def link goSnarkCircuitOut   		Special
hi def link goSnarkCircuitPrivatePublic		Keyword

let b:current_syntax = "go-snark-circuit"
if main_syntax == 'go-snark-circuit'
  unlet main_syntax
endif
let &cpo = s:cpo_save
unlet s:cpo_save

" vim: ts=8
