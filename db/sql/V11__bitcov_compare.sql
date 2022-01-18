CREATE OR REPLACE FUNCTION bitcov_compare(data1 bytea, data2 bytea)
	RETURNS REAL
AS $$
from PIL import Image
import io
import numpy as np

img_array1 = np.array(Image.open(io.BytesIO(data1)))
img_array2 = np.array(Image.open(io.BytesIO(data2)))

if img_array1.size != img_array2.size:
	return -1

misses = 0
for a,b in np.nditer([img_array1, img_array2]):
	if a != b:
		misses = misses + 1

return 1 - (misses / img_array1.size)

$$ LANGUAGE plpython3u;