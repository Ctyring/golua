---
--- Generated by Luanalysis
--- Created by Lenovo.
--- DateTime: 2022/12/28 19:08
---
-- 在lua中没有类的概念
-- 在lua中可以参考基于原型的语言中的一些做法来模拟类(比如javascript)
-- 在这些语言中，对象不属于类，而是对象可以有一个原型，原型也是一种普通对象，当操作在对象上找不到时，会在原型上查找
-- 例如有两个对象AB，让B成为A的一个原型
--A = {}
--B = {x = 1}
--setmetatable(A, {__index = B})
--print(A.x) -- 1
Account = require("object")
a = Account:new{balance = 0}
a:deposit(100.00)
print(a.balance)