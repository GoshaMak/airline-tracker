from collections import defaultdict
from typing import Tuple
import matplotlib.pyplot as plt


def read_avg_float_sums(file_path: str) -> tuple[list[int], list[float]]:
    values = defaultdict(list)

    with open(file_path, "r", encoding="utf-8") as file:
        for line_num, line in enumerate(file, start=1):
            line = line.strip()

            if not line:
                continue

            parts = line.split(":")
            if len(parts) != 3:
                raise ValueError(f"Invalid line {line_num} in {file_path}: {line}")

            x = int(parts[0])
            first_float = float(parts[1])
            second_float = float(parts[2])

            values[x].append(first_float + second_float)

    if not values:
        raise ValueError(f"File does not contain valid data: {file_path}")

    x_values = sorted(values.keys())
    y_values = [sum(values[x]) / len(values[x]) for x in x_values]

    # x = np.array(x_values)
    # y = np.array(y_values)

    # coef = np.polyfit(x, y, deg=1)
    # poly = np.poly1d(coef)
    #
    # x_new = np.linspace(x.min(), x.max(), 300)
    # y_new = poly(x_new)

    return x_values, y_values


def maxImpovement(a, b) -> Tuple[float, float]:
    k = a[0] / b[0]
    sk = k
    for i in range(len(a)):
        k = max(k, a[i] / b[i])
        sk += a[i] / b[i]
    return (k, sk / len(a))


def plot_avg_float_sums(
    first_file_path: str,
    second_file_path: str,
    third_file_path: str,
    first_label: str,
    second_label: str,
    third_label: str,
    save_path: str | None = None,
) -> None:
    first_x, first_y = read_avg_float_sums(first_file_path)
    second_x, second_y = read_avg_float_sums(second_file_path)
    third_x, third_y = read_avg_float_sums(third_file_path)

    print("B-Tree: ", maxImpovement(third_y, first_y))
    print("Hash: ", maxImpovement(third_y, second_y))

    plt.figure(figsize=(10, 6))

    plt.plot(first_x, first_y, linestyle="--", label=first_label)
    plt.plot(second_x, second_y, linestyle=":", label=second_label)
    plt.plot(third_x, third_y, label=third_label)

    plt.xlabel("Размер, шт.")
    plt.ylabel("Время, мс")
    plt.grid(True)
    plt.legend()

    if save_path is not None:
        plt.savefig(save_path, dpi=300, bbox_inches="tight")
    else:
        plt.show()


plot_avg_float_sums(
    "res/flights_b.txt",
    "res/flights_h.txt",
    "res/flights.txt",
    "B-дерево индекс",
    "Хеш индекс",
    "Без индекса",
    "res/flights_compare.pdf",
)
