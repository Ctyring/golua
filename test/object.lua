---
--- Generated by Luanalysis
--- Created by Lenovo.
--- DateTime: 2022/12/28 16:10
---
--Account = {balance = 0}
--function Account:Withdraw(v)
--    self.balance = self.balance - v
--end

-- 调用
--a = Account
--a:Withdraw(100.00)
--a.Withdraw(a, 100.00)
-- 使用参数self是所有面向对象语言的核心点，大多数面向对象语言都向程序员隐藏了这个机制
-- 在lua中也可以通过语法糖:从而不必显式声明这个参数
Account = {
    balance = 0,
    Withdraw = function(self, v)
        self.balance = self.balance - v
    end
}
function Account:deposit(v)
    self.balance = self.balance + v
end
function Account:new(o)
    o = o or {} -- 如果没有传入参数，就创建一个空表
    self.__index = self -- 使Account的实例可以访问Account的方法
    setmetatable(o, self)
    return o
end
return Account
