import json
import pandas as pd
from datetime import datetime
import argparse
import sys
import os
import plotly.express as px
import plotly.graph_objects as go
from plotly.subplots import make_subplots

def parse_envoy_logs(file_path):
    """Парсит файл с логами Envoy в формате JSON"""
    logs = []
    
    with open(file_path, 'r', encoding='utf-8') as file:
        for line in file:
            line = line.strip()
            if line:
                try:
                    log_entry = json.loads(line)
                    logs.append(log_entry)
                except json.JSONDecodeError as e:
                    print(f"Ошибка парсинга JSON: {e}")
                    continue
    
    return logs

def prepare_latency_data(logs):
    """Подготавливает данные для анализа latency"""
    data = []
    
    for log in logs:
        try:
            timestamp = datetime.fromisoformat(log['timestamp'].replace('Z', '+00:00'))
            
            duration_ms = log.get('duration_ms', '0')
            if duration_ms == '-' or duration_ms is None:
                duration_ms = 0
            else:
                duration_ms = float(duration_ms)
            
            # Очищаем path от параметров запроса для группировки
            path = log.get('path', '')
            if '?' in path:
                path = path.split('?')[0]
            
            data.append({
                'timestamp': timestamp,
                'duration_ms': duration_ms,
                'response_code': log.get('response_code', ''),
                'method': log.get('method', ''),
                'path': path,
                'user_agent': log.get('user_agent', ''),
                'authority': log.get('authority', '')
            })
            
        except (ValueError, KeyError) as e:
            continue
    
    return data

def create_latency_line_chart(data, output_file="latency_lines.html"):
    """Создает линейный график latency с временной осью и выделенными цветами"""
    
    if not data:
        print("Нет данных для построения графика")
        return
    
    df = pd.DataFrame(data)
    df = df.sort_values('timestamp')
    
    # Статистика
    total_requests = len(df)
    unique_paths = df['path'].nunique()
    response_codes = sorted(df['response_code'].unique())
    
    print(f"📊 Статистика данных:")
    print(f"   Всего запросов: {total_requests}")
    print(f"   Уникальных путей: {unique_paths}")
    print(f"   Коды ответа: {', '.join(response_codes)}")
    
    # Создаем основной график
    fig = go.Figure()
    
    # Яркая цветовая схема для лучшего различения
    colors = [
        '#FF6B6B', '#4ECDC4', '#45B7D1', '#96CEB4', '#FFEAA7', '#DDA0DD',
        '#98D8C8', '#F7DC6F', '#BB8FCE', '#85C1E9', '#F8C471', '#82E0AA',
        '#F1948A', '#85C1E9', '#D7BDE2', '#F9E79F', '#A9DFBF', '#F5B7B1',
        '#AED6F1', '#D2B4DE', '#FAD7A0', '#ABEBC6', '#F5CBA7', '#AED6F1',
        '#E8DAEF', '#FDEBD0', '#D1F2EB', '#FADBD8', '#D6EAF8', '#EBDEF0'
    ]
    
    # Группируем данные по response_code и path
    grouped_data = df.groupby(['response_code', 'path'])
    
    color_index = 0
    legend_added = set()
    
    for (response_code, path), group in grouped_data:
        if len(group) > 1:  # Только группы с более чем одной точкой для линий
            group = group.sort_values('timestamp')
            
            # Создаем уникальный ключ для легенды
            legend_key = f"{response_code} - {path}"
            
            fig.add_trace(
                go.Scatter(
                    x=group['timestamp'],
                    y=group['duration_ms'],
                    mode='lines+markers',
                    name=legend_key,
                    line=dict(
                        color=colors[color_index % len(colors)],
                        width=3,
                        shape='spline'  # Сглаженные линии
                    ),
                    marker=dict(
                        color=colors[color_index % len(colors)],
                        size=6,
                        opacity=0.8,
                        line=dict(width=1, color='white')
                    ),
                    hovertemplate=(
                        "<b>Time: %{x}</b><br>" +
                        "Latency: %{y:.2f} ms<br>" +
                        f"Code: {response_code}<br>" +
                        f"Path: {path}<br>" +
                        f"Method: {group.iloc[0]['method']}<br>" +
                        "<extra></extra>"
                    ),
                    legendgroup=legend_key,
                    showlegend=legend_key not in legend_added
                )
            )
            legend_added.add(legend_key)
            color_index += 1
    
    # Обновляем layout
    fig.update_layout(
        title=dict(
            text=f"📈 Latency Over Time by Response Code and Path<br>"
                 f"<sub>Total: {total_requests} requests | {unique_paths} unique paths | Codes: {', '.join(response_codes)}</sub>",
            x=0.5,
            font=dict(size=20)
        ),
        height=800,
        xaxis_title="Time",
        yaxis_title="Latency (ms) - Logarithmic Scale",
        showlegend=True,
        hovermode="closest",
        template="plotly_dark",  # Темная тема для лучшего контраста
        legend=dict(
            orientation="v",
            yanchor="top",
            y=1,
            xanchor="left",
            x=1.02,
            font=dict(size=10),
            bgcolor='rgba(0,0,0,0.5)'
        ),
        plot_bgcolor='rgba(0,0,0,0.1)',
        paper_bgcolor='rgba(0,0,0,0.1)'
    )
    
    # Устанавливаем логарифмическую шкалу
    fig.update_yaxes(type="log", gridcolor='rgba(255,255,255,0.1)')
    fig.update_xaxes(gridcolor='rgba(255,255,255,0.1)')
    
    # Сохраняем как HTML
    fig.write_html(
        output_file,
        include_plotlyjs='cdn',
        config={
            'displayModeBar': True,
            'scrollZoom': True,
            'responsive': True
        }
    )
    
    return output_file

def create_faceted_line_charts(data, output_file="latency_faceted_lines.html"):
    """Создает раздельные линейные графики по response_code"""
    
    df = pd.DataFrame(data)
    df = df.sort_values('timestamp')
    
    # Создаем facet grid по кодам ответа с линиями
    fig = px.line(
        df,
        x='timestamp',
        y='duration_ms',
        color='path',
        facet_col='response_code',
        facet_col_wrap=3,
        log_y=True,
        title='Latency Over Time by Response Code and Path (Faceted View)',
        labels={
            'timestamp': 'Time',
            'duration_ms': 'Latency (ms) - Log Scale',
            'path': 'API Path'
        },
        height=900
    )
    
    # Улучшаем внешний вид линий
    fig.update_traces(
        line=dict(width=2.5),
        marker=dict(size=4, opacity=0.8)
    )
    
    fig.update_layout(
        showlegend=True,
        legend=dict(
            orientation="v",
            yanchor="top",
            y=1,
            xanchor="left",
            x=1.02
        ),
        template="plotly_dark"
    )
    
    # Сохраняем
    fig.write_html(output_file, include_plotlyjs='cdn')
    return output_file

def create_animated_latency_chart(data, output_file="latency_animated.html"):
    """Создает анимированный график latency"""
    
    df = pd.DataFrame(data)
    df = df.sort_values('timestamp')
    
    # Создаем анимированный scatter plot
    fig = px.scatter(
        df,
        x='timestamp',
        y='duration_ms',
        color='response_code',
        size='duration_ms',
        hover_name='path',
        animation_frame=df['timestamp'].dt.strftime('%Y-%m-%d %H:%M'),
        log_y=True,
        title='Animated Latency Over Time',
        labels={
            'timestamp': 'Time',
            'duration_ms': 'Latency (ms) - Log Scale',
            'response_code': 'Response Code'
        },
        height=700
    )
    
    fig.update_layout(
        template="plotly_dark",
        showlegend=True
    )
    
    fig.write_html(output_file, include_plotlyjs='cdn')
    return output_file

def create_aggregated_trends(data, output_file="latency_trends.html"):
    """Создает график агрегированных трендов по времени"""
    
    df = pd.DataFrame(data)
    df = df.sort_values('timestamp')
    
    # Агрегируем по временным интервалам
    df['time_window'] = df['timestamp'].dt.floor('1min')  # 1-минутные интервалы
    
    # Создаем агрегированные данные
    aggregated = df.groupby(['time_window', 'response_code', 'path']).agg({
        'duration_ms': ['mean', 'count']
    }).round(2)
    
    aggregated.columns = ['mean_latency', 'request_count']
    aggregated = aggregated.reset_index()
    
    # Фильтруем только значительные группы
    significant_data = aggregated[aggregated['request_count'] >= 1]
    
    fig = px.line(
        significant_data,
        x='time_window',
        y='mean_latency',
        color='path',
        line_dash='response_code',
        log_y=True,
        title='Aggregated Latency Trends by Response Code and Path',
        labels={
            'time_window': 'Time',
            'mean_latency': 'Average Latency (ms) - Log Scale',
            'path': 'API Path',
            'response_code': 'Response Code'
        },
        height=700
    )
    
    fig.update_layout(
        template="plotly_dark",
        showlegend=True,
        legend=dict(
            orientation="v",
            yanchor="top",
            y=1,
            xanchor="left",
            x=1.02
        )
    )
    
    fig.write_html(output_file, include_plotlyjs='cdn')
    return output_file

def print_detailed_statistics(data):
    """Выводит детальную статистику в консоль"""
    
    df = pd.DataFrame(data)
    
    print("\n" + "="*80)
    print("📈 ДЕТАЛЬНАЯ СТАТИСТИКА ПО RESPONSE_CODE И PATH")
    print("="*80)
    
    # Статистика по response_code
    print(f"\n🔴 СТАТИСТИКА ПО КОДАМ ОТВЕТА:")
    code_stats = df.groupby('response_code').agg({
        'duration_ms': ['count', 'mean', 'median', 'min', 'max'],
        'path': 'nunique'
    }).round(2)
    
    code_stats.columns = ['Count', 'Mean Latency', 'Median Latency', 'Min Latency', 'Max Latency', 'Unique Paths']
    
    for code, row in code_stats.iterrows():
        print(f"   Code {code}:")
        print(f"      Requests: {row['Count']}")
        print(f"      Unique paths: {row['Unique Paths']}")
        print(f"      Latency - Mean: {row['Mean Latency']}ms, Median: {row['Median Latency']}ms")
        print(f"      Range: {row['Min Latency']}ms - {row['Max Latency']}ms")
    
    # Топ путей по количеству запросов
    print(f"\n📊 ТОП-10 САМЫХ ЧАСТЫХ ПУТЕЙ:")
    path_stats = df['path'].value_counts().head(10)
    for path, count in path_stats.items():
        path_data = df[df['path'] == path]
        avg_latency = path_data['duration_ms'].mean()
        codes = path_data['response_code'].unique()
        print(f"   {path}:")
        print(f"      Requests: {count}, Avg Latency: {avg_latency:.2f}ms")
        print(f"      Response codes: {', '.join(codes)}")

def main():
    """Основная функция"""
    parser = argparse.ArgumentParser(description='Создание линейных графиков latency с разделением по response_code и path')
    parser.add_argument('file', help='Путь к файлу с логами Envoy')
    parser.add_argument('--output', '-o', default='latency_lines.html', 
                       help='Путь для сохранения основного HTML отчета')
    parser.add_argument('--all', '-a', action='store_true',
                       help='Создать все дополнительные графики и отчеты')
    parser.add_argument('--stats', '-s', action='store_true',
                       help='Показать детальную статистику в консоли')
    
    args = parser.parse_args()
    
    try:
        print(f"📁 Чтение логов из файла: {args.file}")
        logs = parse_envoy_logs(args.file)
        
        if not logs:
            print("❌ Не удалось прочитать логи из файла")
            return
        
        print(f"✅ Успешно прочитано {len(logs)} записей")
        
        data = prepare_latency_data(logs)
        
        if not data:
            print("❌ Нет корректных данных для анализа")
            return
        
        # Всегда показываем базовую статистику
        print_detailed_statistics(data)
        
        print("🔄 Создание основного линейного графика...")
        output_file = create_latency_line_chart(data, args.output)
        
        if args.all:
            print("🔄 Создание дополнительных графиков...")
            faceted_file = create_faceted_line_charts(data, "latency_faceted_lines.html")
            trends_file = create_aggregated_trends(data, "latency_trends.html")
            animated_file = create_animated_latency_chart(data, "latency_animated.html")
            
            print(f"✅ Дополнительные графики созданы:")
            print(f"   - Faceted lines: {faceted_file}")
            print(f"   - Aggregated trends: {trends_file}")
            print(f"   - Animated: {animated_file}")
        
        print(f"✅ Основной график успешно создан: {output_file}")
        print(f"📊 Откройте файл в браузере для просмотра")
        
    except FileNotFoundError:
        print(f"❌ Файл не найден: {args.file}")
    except Exception as e:
        print(f"❌ Ошибка: {e}")

if __name__ == "__main__":
    main()