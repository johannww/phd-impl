# TODOs

- GetStateByRangeWithPagination only works for read only transactions.
    - GetStateByRange works for read/write, but it is capped by "totalQueryLimit" in core.yaml.
    - Thus, I will ask for the keys in range.
    - If the number of keys is equal to the limit, I will ask perform a search using the last key.
    - How do I retrieve totalQueryLimit?

