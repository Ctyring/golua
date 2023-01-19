---
--- Generated by Luanalysis
--- Created by Lenovo.
--- DateTime: 2023/1/18 15:20
---
co = coroutine.wrap(function()
    print("Started")
    for i=1,5 do
        print("yielding")
        coroutine.yield()
    end
    print("Resumed")
end)

co() -- prints "Started"
co() -- prints "yielding"
co() -- prints "yielding"
co() -- prints "yielding"
co() -- prints "yielding"
co() -- prints "yielding"
--co() -- cannot resume dead coroutine
