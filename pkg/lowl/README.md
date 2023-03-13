# lowl

This package implements an assembler for LOWL.

The LOWL specification is available [here](http://www.ml1.org.uk/implementation.html).

## Usage
TODO: Document.

## Output
The output is a Go package that depends only on the standard library.

## Virtual Machine
This package creates assembly code for the virtual machine described in the LOWL documentation.
Key features 

The basic word size is 16 bits.

### Registers
The VM has three registers, A, B, and C.

1. _A_ is the numerical accumulator.
2. _B_ is the index register.
3. _C_ is the character register.

### Memory
Memory consists of 65,536 words.
Each word holds an unlimited number of 16-bit integers.

## Instructions

### Arguments
1. _V_ means a variable name.
1. _N_ means a non-negative decimal integer constant (possibly represented by a name).
1. _OF_ means a call of the OF macro.
1. _N-OF_ means either N or OF.
1. _table label_ means a label attached to a table item.
1. _charname_ means one of the names representing literal character constants (e.g. NLREP).
1. _character_ means a single character.
1. _characters_ means a string of one or more characters.
1. _(A)(B)_ means either A or B.

### Data types
Character (single character), number (may be integer value or pointer).

### Variables
Represented by identifiers.
No character variables.

### Constants
1. _Numerical_: decimal integer or call of _OF_ macro.
1. _Character_: single character in quotes, or name.

### Registers
Three: A, B and C.

### Labels
Represented by identifiers.
Enclosed in square brackets where placed.

### Subroutines
Names are identifiers.
At most one argument.

## _OF_ Macro
The OF macro takes the form `OF(argument)` where _argument_ is one of the following:

1. _N_ * _S_ + _S_
1. _N_ * _S_ - _S_
1. _N_ * _S_
1. _S_ + _S_
1. _S_ - _S_
1. _S_

And _N_ stands for any positive integer, _S_ for any _submacro_, and the `*`, `+` and `-` are multiplication, addition, and subtraction.

### OpCode Table
Arguments have some meaning:
1. (A) is normally an unsigned address.
2. (P) means register _A_ must be preserved.
2. (R) means the operation may be redundant.
3. (X) means one of the following:
   1. A signed integer.
   2. Register _A_ should be clobbered.
   3. There is no parameter for a subroutine.

The op codes are:

    OpCode Splat Arguments_________________ Meaning______________________________________________
    AAL          N-OF                       add to A a literal.
    AAV          V                          add to A a variable.
    ABV          V                          add to B a variable.
    ALIGN                                   align A up to next boundary.
    ANDL         N                          "and" A with a literal.
    ANDV         V                          "and" A with a variable.
    BMOVE    *                              backwards block move.
    BSTK     *                              stack A on backwards stack.
    BUMP     *   V,N-OF                     increase a variable.
    CAI      *   V,(X)                      compare A indirect signed integer.
    CAI      *   V,(A)                      compare A with indirect address.
    CAL      *   N-OF                       compare A with literal.
    CAV      *   V,(X)                      compare A with variable signed integer.
    CAV      *   V,(A)                      compare A with address.
    CCI      *   V                          compare C indirect.
    CCL      *   'character'                compare C with literal.
    CCN      *   charname                   compare C with named character.
    CFSTK    *                              stack C on forwards stack.
    CLEAR    *   V                          set variable to zero.
    CON          N-OF                       numerical constant.
    CSS      *                              clear subroutine stack (if any).
    DCL          V                          declare variable.
    EQU          V,V                        equate two variables.
    EXIT     *   N,subroutine name          exit from subroutine.
    FMOVE    *                              forwards block move.
    FSTK     *                              stack A on forwards stack.
    GO           label spec                 unconditional branch.
    GOADD    *   V                          multi-way branch.
    GOEQ     *   label spec                 branch if equal.
    GOGE     *   label spec                 branch if greater than or equal.
    GOGR     *   label spec                 branch if greater than.
    GOLE     *   label spec                 branch if less than or equal.
    GOLT     *   label spec                 branch if less than.
    GOND     *   label spec                 branch if C is not a digit; otherwise put value in A.
    GONE     *   label spec                 branch if not equal.
    GOPC     *   label spec                 branch if C is a punctuation character.
    GOSUB        subroutine name,(distance) call subroutine.
    GOSUB        subroutine name,(X)        call subroutine.
    IDENT        V,decimal integer          equate name to integer.
    LAA          V,D                        load A modified (variable).
    LAA          table label,C              load A modified (table item).
    LAI          V,(R)                      load A indirect.
    LAI          V,(X)                      load A indirect.
    LAL          N-OF                       load A with literal.
    LAM          N-OF                       load A modified.
    LAV          V,(R)                      load A with variable.
    LAV          V,(X)                      load A with variable.
    LBV          V                          load B with variable.
    LCI          V,(R)                      load C indirect.
    LCI          V,(X)                      load C indirect.
    LCM          N-OF                       load C modified.
    LCN          charname                   load C with named character.
    MESS     *   'characters'               output a message.
    MULTL        N-OF                       multiply A by a literal.
    NB           'characters'               comment.
    NCH          charname                   character constant.
    PRGEN                                   end of logic.
    PRGST        'characters'               start of logic.
    SAL          N-OF                       subtract from A a literal.
    SAV          V                          subtract from A a variable.
    SBL          N-OF                       subtract from B a literal.
    SBV          V                          subtract from B a variable.
    STI          V,(P)                      store A indirectly in variable.
    STI          V,(X)                      store A indirectly in variable.
    STR          'characters'               character string constant.
    STV          V,(P)                      store A in variable.
    STV          V,(X)                      store A in variable.
    SUBR         subroutine name,(PARNM),N  declare subroutine.
    SUBR         subroutine name,(X)        declare subroutine.
    UNSTK    *   V                          unstack from backwards stack.

# Source Licenses

LOWL and L code and documentation are

   Copyright (c) 1972,2023 P.J. Brown, R.D. Eager. All rights reserved.
   
   Permission is granted to copy and/or modify documentation/code for private
   use only.  Machine readable versions must not be placed on public web sites
   or FTP sites, or otherwise made generally accessible in an electronic form.
   Instead, please provide a link to the original documentation/code on the
   official ML/I web site (http://www.ml1.org.uk).
