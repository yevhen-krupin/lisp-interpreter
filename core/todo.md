1. no panic: errors should be returned, without crashing cli or server or test
2. find all possible errors in
    a. tokenizer (e.g. braces consistency)
    b. parser (ast) - whatever we already throw, maybe some more
    c. consolidate the errors with RI for same situation
3. fix (fact 100) so there is no overload (big num)