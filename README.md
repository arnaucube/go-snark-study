
## Caution, Warning
Fork UNDER CONSTRUCTION! Will ask for merge soon


Current implementation status:
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
[[0 0 210 0 0 0 0 0 0 0 0 0] [0 0 0 1 0 0 0 0 0 0 0 0] [0 0 0 1 0 0 0 0 0 0 0 0] [0 210 0 0 0 0 0 0 0 0 0 0] [0 0 0 0 0 0 1 0 0 0 0 0] [0 0 0 0 0 0 1 0 0 0 0 0] [0 1 0 0 0 0 0 0 0 0 0 0] [0 1 0 0 0 0 0 0 0 0 0 0] [1 0 0 0 0 0 0 0 0 0 0 0]]
[[0 0 210 0 0 0 0 0 0 0 0 0] [0 0 5 0 0 0 0 0 0 0 0 0] [0 0 0 0 1 0 0 0 0 0 0 0] [0 210 0 0 0 0 0 0 0 0 0 0] [0 5 0 0 0 0 0 0 0 0 0 0] [0 0 0 0 0 0 0 1 0 0 0 0] [0 1 0 0 0 0 0 0 0 0 0 0] [0 0 0 0 0 0 0 0 0 1 0 0] [0 0 0 0 0 1 0 0 1 0 1 0]]
[[0 0 0   1 0 0 0 0 0 0 0 0] [0 0 0 0 1 0 0 0 0 0 0 0] [0 0 0 0 0 1 0 0 0 0 0 0] [0 0   0 0 0 0 1 0 0 0 0 0] [0 0 0 0 0 0 0 1 0 0 0 0] [0 0 0 0 0 0 0 0 1 0 0 0] [0 0 0 0 0 0 0 0 0 1 0 0] [0 0 0 0 0 0 0 0 0 0 1 0] [0 0 0 0 0 0 0 0 0 0 0 1]]
input
[7 11]
witness
[1 7 11 5336100 293485500 1566067976550000 2160900 75631500 163432108350000 49 343 1729500084900343]
```
Note that we only need 9 multiplication Gates instead of 16
