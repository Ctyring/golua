---
--- Generated by Luanalysis
--- Created by Lenovo.
--- DateTime: 2023/3/17 16:11
---
print("Hello World!")

-- 阶乘
function fact(n)
    if n == 0 then
        return 1
    else
        return n * fact(n - 1)
    end
end

print(fact(5))

--dofile("test/test2.lua")

-- 基本类型
print(type(nil)) -- nil
print(type(true)) -- boolean
print(type(10.4 * 3)) -- number
print(type("hello world")) -- string
print(type(print)) -- function
print(type(type)) -- function
print(type({})) -- table
print(type(type(X))) -- string 不管X是什么，type都会返回字符串，所以是string

print(arg)