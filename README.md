
## Caution, Warning
Fork UNDER CONSTRUCTION! Will ask for merge soon


Current implementation status:
- [x] optimized gate reduction!! Reusing gates as often as possible! See the awesome results below :)
- [x] extended circuit code compiler
- [x] move witness calculation outside the setup phase 
- [x] fixed hard bugs

### Library usage
Warning: not finished.

Working example of gate-reduction and code parsing:
```go
def do(x):
    e = x * 5
    b = e * 6
    c = b * 7
    f = c * 1
    d = c * f
    out = d * mul(d,e)

def doSomethingElse(x ,k):
    z = k * x
    out = do(x) + mul(x,z)

def main(x,z):
    out = do(z) + doSomethingElse(x,x)

def mul(a,b):
    out = a * b
```
R1CS Output:
```go
[[0 0 1 0 0 0 0 0 0 0] [0 0 0 0 1 0 0 0 0 0] [0 0 0 0 1 0 0 0 0 0] [0 1 0 0 0 0 0 0 0 0] [0 0 0 0 0 0 0 1 0 0] [0 0 0 0 0 0 0 1 0 0] [1 0 0 0 0 0 0 0 0 0]]
[[0 0 1 0 0 0 0 0 0 0] [0 0 1 0 0 0 0 0 0 0] [0 0 0 0 0 1 0 0 0 0] [0 1 0 0 0 0 0 0 0 0] [0 1 0 0 0 0 0 0 0 0] [0 0 0 0 0 0 0 0 1 0] [0 0 0 0 0 0 9724050000 0 1 9724050000]]
[[0 0 0 0 1 0 0 0 0 0] [0 0 0 0 0 1 0 0 0 0] [0 0 0 0 0 0 1 0 0 0] [0 0 0 0 0 0 0 1 0 0] [0 0 0 0 0 0 0 0 1 0] [0 0 0 0 0 0 0 0 0 1] [0 0 0 1 0 0 0 0 0 0]]
input
[7 11]
witness
[1 7 11 1729500084900343 121 1331 161051 49 343 16807]
another input
[365235 11876525]
witness
[1 365235 11876525 2297704271284150716235246193843898764109352875 141051846075625 1675205776213312203125 236290867291438012851239954111328125 133396605225 48721109109352875 6499230557984496821593771875]

```
Note that we only need 7 multiplication Gates instead of 16. The 4th witness value is the programs output. Use python script to check correctness!
