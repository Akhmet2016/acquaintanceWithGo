/*package main

import "fmt"

func main() {
	func () { // Объявление анонимной функции
		fmt.Println("Анонимная функция") // Тело функции
	}() // Закрытие и вызов самой себя
	inc := increment() // Объявляем функцию инкремент
	fmt.Println(inc()) // Будет 1
	fmt.Println(inc()) // Будет 2 Особенности области видимости, без замыкания постоянно выводилась бы 1
}

func increment() func() int {
	count := 0 // Особенность области видимости, при возвращение будет увеличиваться на единицу
	return func() int { // Возвращаем функцию
		count++  // Тело функции
		return count
	} // не вызываем саму себя
}*/

package main

func main() {
	
}
