#!/usr/bin/python
# encoding=UTF-8

numbers = [12, 37, 5, 42, 8, 3]
even = []
odd = []

while len(numbers) > 0:
    number = numbers.pop()
    if number % 2 == 0:
        even.append(number)
    else:
        odd.append(number)

print even
print odd

count = 0
while (count < 9):
    count = count + 1
    if count % 6 == 0:
        break
    if count % 2 == 0:
        continue
    print 'The count is', count
else:
    print "while is gone"

print "bye"

for letter in 'Python':
    print "letter", letter
fruits = ['banana', 'apple', 'mango']
for f in fruits:
    print 'fruit', f

for index in range(len(fruits)):
    print "fruit: ", fruits[index]

for i in range(10):
    print i,

for i in range(10, 20):
    for n in range(2, i):
        if i % n == 0:
            j = i / n
            print "%d*%d=%d" % (i, n, j)
            break
    else:
        print "z质数 ", i
