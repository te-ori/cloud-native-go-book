// # Stability patterns
//
// Bu patternler herhangi bir başka servise istekte bulunulduğunda
// herhangi bir sebepten dolayı bu istek hataya neden olursa bu hatanın
// hem sistemin geriye kalanının çalışmasını engellememeye hem de istekte
// bulunan kullancının ve sistemin geriye kalanının hata ve sonuçları
// konusunda sağılklı bir şekilde bilgiliendirilmelerini sağar
package stability_patterns

import "context"

type Circuit func(context.Context) (string, error)
