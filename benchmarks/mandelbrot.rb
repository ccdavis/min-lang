#!/usr/bin/env ruby

WIDTH = 80
HEIGHT = 40
MAX_ITER = 100
XMIN = -2.5
XMAX = 1.0
YMIN = -1.0
YMAX = 1.0

def mandelbrot(cx, cy)
  x = 0.0
  y = 0.0
  iter = 0

  while iter < MAX_ITER
    x2 = x * x
    y2 = y * y

    return iter if x2 + y2 > 4.0

    xtemp = x2 - y2 + cx
    y = 2.0 * x * y + cy
    x = xtemp

    iter += 1
  end

  MAX_ITER
end

def iter_to_char(iter)
  return " " if iter == MAX_ITER
  return "." if iter > 80
  return ":" if iter > 60
  return "-" if iter > 40
  return "=" if iter > 20
  return "+" if iter > 10
  return "*" if iter > 5
  "#"
end

def main
  puts "Rendering Mandelbrot Set..."
  puts "Size: #{WIDTH} x #{HEIGHT}"
  puts "Max iterations: #{MAX_ITER}"
  puts ""

  (0...HEIGHT).each do |row|
    line = ""
    (0...WIDTH).each do |col|
      cx = XMIN + (XMAX - XMIN) * col / WIDTH.to_f
      cy = YMIN + (YMAX - YMIN) * row / HEIGHT.to_f

      iter = mandelbrot(cx, cy)
      char = iter_to_char(iter)
      line += char
    end

    puts line
  end

  puts ""
  puts "Rendering complete!"
end

main if __FILE__ == $0
