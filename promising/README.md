# Promising

## Why you should consider this package
Usually, a cache implementation would provide a `Fetch` method which takes in a `key` and a `fn`, which gets
executed if the given `key` does not exist in the cache. After that, the result of this `fn` would be saved
into the cache to serve subsequent `Fetch` with the same `key`.

This cache implementation is different in that it does not accept any `fn`. Instead, when clients attempt to
fetch a result using a `key`, they will get back a `promise`. Subsequently, clients can block and wait for the
promise to resolve. 

Wait a minute... If clients are not populating the cache, who does? Great question! On the other end of this
cache is usually a batch processor. For example, in my company, we have a process to load people using phone
numbers. Initially, this batch processor is given a list of phone numbers to start loading. Subsequently, it 
will recursively go through each person's contact list to load more people. As you can imagine, when we first
start loading, other than the seed phone numbers, we don't actually know if our code would come into contact
with any other phone numbers.

This is where the `promising` nature of this implementation comes from. When our batch processor starts loading,
it will keep populating the cache of people using phone number as key. A client can attempt to fetch a person
using a phone number to get back a promise. When our batch processor completes, all promises will be resolved.
Waiting clients will either get back a valid person or an error saying the person they're looking for does not
exist at all.

After our batch processor completes loading, no further results will be coming as the processor no longer puts
anything into our cache. Hence, when a new client tries to fetch a person, they will get back 1 of these results: 
1. a valid person if the given phone number exists in the cache
2. an error saying no further people will be arriving in the cache

## Sample usage

```go
cache := promising.NewCache[string, int]()

go func() {
    // Signal that no further results will arrive
    defer cache.Destroy()
    
    for i := 0 ; i < 10 ; i++ {
        key := fmt.Sprintf("key%v", i)
        
        if i % 2 == 0 {
            cache.Complete(key, 0, assert.AnError)
            continue
        }
        
        cache.Complete(key, i, nil)
    }
}()

for i := 0 ; i < 20 ; i++ {
    go func() {
        key := fmt.Sprintf("key%v", i)
        
        result, err := cache.GetOutcome(key).Get(context.Background())
        fmt.Println(key)
        fmt.Println(result)
        fmt.Println(err)
    }()
}
```
