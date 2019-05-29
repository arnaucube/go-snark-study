
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
[[0 0 210 0 0 0 0 0 0 0 0 0 0 0] [0 0 210 0 0 0 0 0 0 0 0 0 0 0] [0 0 0 0 1 0 0 0 0 0 0 0 0 0] [0 0 0 1 0 0 0 0 0 0 0 0 0 0] [0 210 0 0 0 0 0 0 0 0 0 0 0 0] [0 210 0 0 0 0 0 0 0 0 0 0 0 0] [0 0 0 0 0 0 0 0 1 0 0 0 0 0] [0 0 0 0 0 0 0 1 0 0 0 0 0 0] [0 1 0 0 0 0 0 0 0 0 0 0 0 0] [0 1 0 0 0 0 0 0 0 0 0 0 0 0] [1 0 0 0 0 0 0 0 0 0 0 0 0 0]]
[[0 0 210 0 0 0 0 0 0 0 0 0 0 0] [0 0 210 0 0 0 0 0 0 0 0 0 0 0] [0 0 5 0 0 0 0 0 0 0 0 0 0 0] [0 0 0 0 0 1 0 0 0 0 0 0 0 0] [0 210 0 0 0 0 0 0 0 0 0 0 0 0] [0 210 0 0 0 0 0 0 0 0 0 0 0 0] [0 5 0 0 0 0 0 0 0 0 0 0 0 0] [0 0 0 0 0 0 0 0 0 1 0 0 0 0] [0 1 0 0 0 0 0 0 0 0 0 0 0 0] [0 0 0 0 0 0 0 0 0 0 0 1 0 0] [0 0 0 0 0 0 1 0 0 0 1 0 1 0]]
[[0 0 0   1 0 0 0 0 0 0 0 0 0 0] [0 0 0   0 1 0 0 0 0 0 0 0 0 0] [0 0 0 0 0 1 0 0 0 0 0 0 0 0] [0 0 0 0 0 0 1 0 0 0 0 0 0 0] [0 0   0 0 0 0 0 1 0 0 0 0 0 0] [0 0   0 0 0 0 0 0 1 0 0 0 0 0] [0 0 0 0 0 0 0 0 0 1 0 0 0 0] [0 0 0 0 0 0 0 0 0 0 1 0 0 0] [0 0 0 0 0 0 0 0 0 0 0 1 0 0] [0 0 0 0 0 0 0 0 0 0 0 0 1 0] [0 0 0 0 0 0 0 0 0 0 0 0 0 1]]
```
Note that we only need 11 multiplication Gates instead of 16
