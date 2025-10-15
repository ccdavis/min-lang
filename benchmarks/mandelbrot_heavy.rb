#!/usr/bin/env ruby

def main
  puts "=== Heavy Mandelbrot Benchmark ==="
  puts "This benchmark is designed to minimize startup time effects"
  puts ""

  # Test 1: Large resolution, high iteration count
  puts "Test 1: 200x200 @ 1000 iterations"
  width1 = 200
  height1 = 200
  max_iter1 = 1000

  total_iterations = 0

  (0...height1).each do |row|
    (0...width1).each do |col|
      cx = -2.5 + 3.5 * col / width1.to_f
      cy = -1.25 + 2.5 * row / height1.to_f

      x = 0.0
      y = 0.0
      iter = 0

      while iter < max_iter1
        x2 = x * x
        y2 = y * y

        break if x2 + y2 > 4.0

        xtemp = x2 - y2 + cx
        y = 2.0 * x * y + cy
        x = xtemp

        iter += 1
      end

      total_iterations += iter
    end
  end

  puts "Pixels calculated: #{width1 * height1}"
  puts "Total iterations: #{total_iterations}"
  puts "Average iterations per pixel: #{total_iterations / (width1 * height1)}"
  puts ""

  # Test 2: Very high iteration count (deep zoom)
  puts "Test 2: 150x150 @ 2000 iterations (deep zoom)"
  width2 = 150
  height2 = 150
  max_iter2 = 2000

  zoom_x = -0.5
  zoom_y = 0.0
  zoom_size = 0.5

  total_iterations = 0

  (0...height2).each do |row|
    (0...width2).each do |col|
      cx = zoom_x - zoom_size + 2.0 * zoom_size * col / width2.to_f
      cy = zoom_y - zoom_size + 2.0 * zoom_size * row / height2.to_f

      x = 0.0
      y = 0.0
      iter = 0

      while iter < max_iter2
        x2 = x * x
        y2 = y * y

        break if x2 + y2 > 4.0

        xtemp = x2 - y2 + cx
        y = 2.0 * x * y + cy
        x = xtemp

        iter += 1
      end

      total_iterations += iter
    end
  end

  puts "Pixels calculated: #{width2 * height2}"
  puts "Total iterations: #{total_iterations}"
  puts "Average iterations per pixel: #{total_iterations / (width2 * height2)}"
  puts ""

  # Test 3: Multiple medium-resolution renders
  puts "Test 3: 30 frames of 100x100 @ 500 iterations"
  frames = 30
  frame_width = 100
  frame_height = 100
  frame_iters = 500

  total_pixels = 0
  total_iterations = 0

  (0...frames).each do |frame|
    zoom_factor = 1.0 - frame * 0.02

    (0...frame_height).each do |row|
      (0...frame_width).each do |col|
        cx = -2.0 * zoom_factor + 3.0 * zoom_factor * col / frame_width.to_f
        cy = -1.0 * zoom_factor + 2.0 * zoom_factor * row / frame_height.to_f

        x = 0.0
        y = 0.0
        iter = 0

        while iter < frame_iters
          x2 = x * x
          y2 = y * y

          break if x2 + y2 > 4.0

          xtemp = x2 - y2 + cx
          y = 2.0 * x * y + cy
          x = xtemp

          iter += 1
        end

        total_iterations += iter
        total_pixels += 1
      end
    end
  end

  puts "Frames calculated: #{frames}"
  puts "Total pixels: #{total_pixels}"
  puts "Total iterations: #{total_iterations}"
  puts "Average iterations per pixel: #{total_iterations / total_pixels}"
  puts ""

  puts "=== Benchmark Complete ==="
end

main if __FILE__ == $0
