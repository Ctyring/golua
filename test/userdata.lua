---
--- Generated by Luanalysis
--- Created by Lenovo.
--- DateTime: 2023/3/18 20:05
---
a = boolarray.new(10)
print(a:size())
a:set(1, true)
print(a:get(1))
print(#a)
print(a)
print(a[1])
a[3] = true
print(a[3])