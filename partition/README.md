# Partition

When you receive a lot of data concurrently, it might be useful to divide the data into partitions asynchronously
before consuming.

```go
partitionFunc := func(a animal) (string, bool) {
    if a.species == "" {
        return "", false
    }
    
    return a.species, true
}

p := NewPartitioner(context.Background(), partitionFunc)

input := []animal{
    {"dog", "name1"},
    {"snail", "name2"},
    {"dog", "name4"},
    {"cat", "name5"},
}

p.Take(input...)

res := p.Outcome()
```

See `partition_test.go` for a detailed example on how to use this feature.