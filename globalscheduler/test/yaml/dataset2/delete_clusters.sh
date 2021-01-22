#!/bin/bash
#!/bin/bash
counter=1
while [ $counter -le 9 ]
do
    name="kubectl delete cluster cluster00$counter"
    echo $name
    $name
     ((counter++))
done
counter=10
while [ $counter -le 99 ]
do
    name="kubectl delete cluster cluster0$counter"
    echo $name
    $name
    ((counter++))
done
counter=100
while [ $counter -le 999 ]
do
    name="kubectl delete cluster cluster$counter"
    $name
    echo $name
    ((counter++))
done
kubectl delete cluster cluster1000